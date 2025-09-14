package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"log"
	"os"
	"strconv"
	"time"

	goldmane "goldmane.golang.com/gapic/cloud.tigera.io/tigera/goldmane/v1"
	goldmanepb "goldmane.golang.com/tigera/goldmane/v1"

	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const GOLDMANE_SERVICE = "goldmane.calico-system.svc"

type statsStream struct {
	Client   *goldmane.StatisticsClient
	TLSCreds credentials.TransportCredentials
}

func (sStream *statsStream) init() {
	endpoint := GOLDMANE_SERVICE + ":7443"
	ctx := context.Background()
	var err2 error

	sStream.Client, err2 = goldmane.NewStatisticsClient(ctx, option.WithoutAuthentication(),
		option.WithEndpoint(endpoint),
		option.WithGRPCDialOption(grpc.WithTransportCredentials(sStream.TLSCreds)),
	)

	if err2 != nil {
		log.Fatalf("failed to create client: %v", err2)
	}
	_ = sStream.Client
}

func main() {
	endpoint := GOLDMANE_SERVICE + ":7443"
	ctx := context.Background()

	// 1. Load Certificates
	clientCert, err := tls.LoadX509KeyPair(os.Getenv("SSL_CERT_PATH"), os.Getenv("SSL_KEY_PATH"))
	if err != nil {
		log.Fatalf("failed to load client key pair: %v", err)
	}

	cert, err := os.ReadFile(os.Getenv("SSL_CERT_PATH"))
	if err != nil {
		log.Fatalf("failed to read CA certificate: %v", err)
	}
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(cert)

	// 2. Create TLS Configuration
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
		// Optional: InsecureSkipVerify: true, for development, but not recommended for production
	}

	sStream := &statsStream{}

	sStream.TLSCreds = credentials.NewTLS(tlsConfig)

	flowsClient, err := goldmane.NewFlowsClient(ctx, option.WithoutAuthentication(),
		option.WithEndpoint(endpoint),
		option.WithGRPCDialOption(grpc.WithTransportCredentials(sStream.TLSCreds)),
	)

	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	sStream.init()

	// Default (empty) request
	flowsReq := &goldmanepb.FlowStreamRequest{}
	clientStream, err := flowsClient.Stream(ctx, flowsReq)
	if err != nil {
		log.Fatalf("flows stream request failed: %v", err)
	}

	ec := &ElasticClient{}
	ec.init()

	go func() {
		for {
			flowr, err := clientStream.Recv()
			if err != nil {
				log.Printf("stream ended: %v", err)
				continue
			}
			//fmt.Printf("Got flow: %+v\n", flowr)
			// Inject flow into Elasticsearch
			ec.InjectFlow(flowr.Flow)
		}
	}()
	seconds, err := strconv.Atoi(os.Getenv("STATS_POLL_TIME"))
	if err != nil {
		log.Fatalf("Error converting string to int: %v", err)
		seconds = 60
	}
	go func() {
		for {
			statisticsReq := &goldmanepb.StatisticsRequest{}
			statisticsStream, err2 := sStream.Client.List(ctx, statisticsReq)
			if err2 != nil {
				log.Fatalf("statistics stream request failed: %v", err2)
			}

			for {
				statsr, err := statisticsStream.Recv()
				if err != nil {
					// Sleep for 3 seconds
					time.Sleep(time.Duration(seconds) * time.Second)
					break
				}
				//fmt.Printf("Got stat: %+v\n", statsr)
				// Inject flow into Elasticsearch
				ec.InjectStatistics(statsr)
			}
		}
	}()

	select {}

}
