package goldmane_cli

import (
	"context"
	"fmt"
	"log"
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

	// Enable SSL (system CAs), no auth
	tlsCreds := credentials.NewClientTLSFromCert(nil, "")

	client, err := goldmane.NewFlowsClient(ctx, option.WithoutAuthentication(),
		option.WithEndpoint(endpoint),
		option.WithGRPCDialOption(grpc.WithTransportCredentials(tlsCreds)),
	)

	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	defer client.Close()

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
