package main

import (
	"bytes"
	"crypto/tls"
	"log"
	"net/http"
	"os"

	"github.com/elastic/go-elasticsearch/v8"
	goldmanepb "goldmane.golang.com/tigera/goldmane/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

const ES_HOST = "https://elasticsearch-sample-es-http.elastic-system.svc:9200"
const ES_FLOWS_INDEX = "goldmane.flows"
const ES_STATS_INDEX = "goldmane.statistics"

var ES_PASS = os.Getenv("ES_PASS")
var ES_USER = os.Getenv("ES_USER")

type ElasticClient struct {
	client *elasticsearch.Client
}

func (ec *ElasticClient) init() {
	// Create Elasticsearch client
	var err error
	ec.client, err = elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{ES_HOST},
		Username:  ES_USER,
		Password:  ES_PASS,
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

func (ec *ElasticClient) InjectStatistics(stats *goldmanepb.StatisticsResult) {
	jsonBytes, err := protojson.Marshal(stats)
	if err != nil {
		log.Fatalf("failed to marshal stats protobuf: %v", err)
	}

	resp, _ := ec.client.Index(
		ES_STATS_INDEX,
		bytes.NewReader(jsonBytes),
	)

	if resp.IsError() {
		log.Fatalf("failed to create Elasticsearch document: %v", resp.Status())
	}

	//log.Printf("Injected stats into Elasticsearch: %+v\n", string(jsonBytes))
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

	resp, _ := ec.client.Index(
		ES_FLOWS_INDEX,
		bytes.NewReader(jsonBytes),
	)

	if resp.IsError() {
		log.Fatalf("failed to create Elasticsearch document: %v", resp.Status())
	}
	//log.Printf("%v", resp)
	//log.Printf("Injected flow into Elasticsearch: %+v\n", string(jsonBytes))
}
