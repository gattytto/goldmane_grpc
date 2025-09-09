package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"

	goldmane "goldmane.golang.com/gapic/cloud.tigera.io/tigera/goldmane/v1"
	goldmanepb "goldmane.golang.com/tigera/goldmane/v1"

	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const GOLDMANE_SERVICE = "goldmane.calico-system.svc"

func main() {
	endpoint := GOLDMANE_SERVICE + ":7443"
	ctx := context.Background()

	// 1. Load Certificates
	clientCert, err := tls.LoadX509KeyPair("../../bundle.pem", "../../goldmane.key")
	if err != nil {
		log.Fatalf("failed to load client key pair: %v", err)
	}

	cert, err := os.ReadFile("../../bundle.pem")
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

	tlsCreds := credentials.NewTLS(tlsConfig)
	client, err := goldmane.NewFlowsClient(ctx, option.WithoutAuthentication(),
		option.WithEndpoint(endpoint),
		option.WithGRPCDialOption(grpc.WithTransportCredentials(tlsCreds)),
	)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
		defer client.Close()
	}

	// Default (empty) request
	req := &goldmanepb.FlowStreamRequest{}

	stream, err := client.Stream(ctx, req)
	if err != nil {
		log.Fatalf("stream request failed: %v", err)

	}
	
    ec := &ElasticClient{}
    ec.init()
	// Read async
	go func() {
		for {
			flowr, err := stream.Recv()
			if err != nil {
				log.Printf("stream ended: %v", err)
				continue
			}
			fmt.Printf("Got flow: %+v\n", flowr)
			// Inject flow into Elasticsearch
			ec.InjectFlow(flowr.Flow)
			continue
		}
	}()
	select {}

}
