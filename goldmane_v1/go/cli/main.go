package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"time"

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

	caCert, err := os.ReadFile("../../bundle.pem")
	if err != nil {
		log.Fatalf("failed to read CA certificate: %v", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// 2. Create TLS Configuration
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      caCertPool,
		// Optional: InsecureSkipVerify: true, for development, but not recommended for production
	}

	// Enable SSL (system CAs), no auth
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

	// Read async
	go func() {
		for {
			flow, err := stream.Recv()
			if err != nil {
				log.Printf("stream ended: %v", err)
				return
			}
			fmt.Printf("Got flow: %+v\n", flow)
		}
	}()

	time.Sleep(10 * time.Second)
}
