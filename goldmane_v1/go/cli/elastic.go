package main

import (
	"bytes"
	"crypto/tls"
	"log"
	"net/http"

	"github.com/elastic/go-elasticsearch/v8"
	goldmanepb "goldmane.golang.com/tigera/goldmane/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

const ES_HOST = "https://elasticsearch-sample-es-http.elastic-system.svc:9200"
const ES_INDEX = "goldmane.flows"

type ElasticClient struct {
	client *elasticsearch.Client
}

func (ec *ElasticClient) init() {
	// Create Elasticsearch client
	var err error
	ec.client, err = elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{ES_HOST},
		Username:  "elastic",
		Password:  "REDACTED",
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	})

	if err != nil {
		log.Fatalf("failed to create Elasticsearch client: %v", err)
	}
}

func (ec *ElasticClient) InjectFlow(flow *goldmanepb.Flow) {

	// // Create Elasticsearch index
	// res, err := client.Indices.Create(
	// 	ES_INDEX,
	// 	client.Indices.Create.WithBody(strings.NewReader(`{
	// 		"settings": {
	// 			"number_of_shards": 1,
	// 			"number_of_replicas": 1
	// 		}
	// 	}`)),
	// )
	// if err != nil {
	// 	log.Fatalf("Error creating index: %s", err)
	// }

	// if res.IsError() {
	// 	log.Printf("Error creating index: %s", res.String())
	// } else {
	// 	log.Println("Index created successfully")
	// }
	// Marshal protobuf message into JSON
	jsonBytes, err := protojson.Marshal(flow)
	if err != nil {
		log.Fatalf("failed to marshal flow protobuf: %v", err)
	}
	// Create Elasticsearch document

	_, err = ec.client.Index(
		ES_INDEX,
		bytes.NewReader(jsonBytes),
	)

	if err != nil {
		log.Fatalf("failed to create Elasticsearch document: %v", err)
	}

	log.Printf("Injected flow into Elasticsearch: %+v\n", string(jsonBytes))
}
