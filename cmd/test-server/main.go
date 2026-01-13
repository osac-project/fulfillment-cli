/*
Copyright (c) 2025 Red Hat Inc.

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the
License. You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific
language governing permissions and limitations under the License.
*/

// This is a test server that simulates the fulfillment service for testing the watch functionality.
// Run this server, then in another terminal run: ./fulfillment-cli get clusters --watch
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	eventsv1 "github.com/innabox/fulfillment-common/api/events/v1"
	ffv1 "github.com/innabox/fulfillment-common/api/fulfillment/v1"
	metadatav1 "github.com/innabox/fulfillment-common/api/metadata/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/innabox/fulfillment-cli/internal/testing"
)

const (
	serverPort          = "8080"
	defaultScenarioFile = "internal/testing/testdata/cluster-lifecycle.yaml"
)

// loggingEventsServer wraps EventsServerFuncs to add logging for the standalone server
type loggingEventsServer struct {
	*testing.EventsServerFuncs
}

func (s *loggingEventsServer) Watch(request *eventsv1.EventsWatchRequest, stream eventsv1.Events_WatchServer) error {
	filter := request.GetFilter()
	log.Printf("Client connected. Filter: %s", filter)

	// Wrap the stream to add logging
	loggingStream := &loggingWatchServer{Events_WatchServer: stream}

	// Call the underlying Watch function
	err := s.EventsServerFuncs.Watch(request, loggingStream)

	if err != nil {
		log.Printf("Client disconnected: %v", err)
	} else {
		log.Println("Client disconnected")
	}

	return err
}

// loggingWatchServer wraps the stream to log sent events
type loggingWatchServer struct {
	eventsv1.Events_WatchServer
}

func (s *loggingWatchServer) Send(response *eventsv1.EventsWatchResponse) error {
	event := response.GetEvent()
	log.Printf("Sending %s event for %s...", event.Type, testing.GetEventObjectID(event))
	return s.Events_WatchServer.Send(response)
}

// Dummy clusters server - just to satisfy the CLI's requirements
type clustersServer struct {
	ffv1.UnimplementedClustersServer
}

func (s *clustersServer) List(ctx context.Context, request *ffv1.ClustersListRequest) (*ffv1.ClustersListResponse, error) {
	// Return empty list
	return &ffv1.ClustersListResponse{}, nil
}

// Dummy metadata server - required for login
type metadataServer struct {
	metadatav1.UnimplementedMetadataServer
}

func (s *metadataServer) Get(ctx context.Context, request *metadatav1.MetadataGetRequest) (*metadatav1.MetadataGetResponse, error) {
	// Return minimal metadata - no authentication required for test server
	return &metadatav1.MetadataGetResponse{}, nil
}

func main() {
	// Parse command line flags
	scenarioFile := flag.String("scenario", defaultScenarioFile, "Path to event scenario YAML file")
	flag.Parse()

	// Load event scenario from file
	scenario, err := testing.LoadScenarioFromFile(*scenarioFile)
	if err != nil {
		log.Fatalf("Failed to load event scenario from %s: %v", *scenarioFile, err)
	}
	log.Printf("Loaded event scenario: %s - %s", scenario.Name, scenario.Description)

	listener, err := net.Listen("tcp", "127.0.0.1:"+serverPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	// Create events server using the builder with loaded scenario
	eventsServerFuncs := testing.NewMockEventsServerBuilder().
		WithScenario(scenario).
		Build()
	eventsv1.RegisterEventsServer(grpcServer, &loggingEventsServer{EventsServerFuncs: eventsServerFuncs})

	ffv1.RegisterClustersServer(grpcServer, &clustersServer{})
	metadatav1.RegisterMetadataServer(grpcServer, &metadataServer{})

	// Register health service
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	reflection.Register(grpcServer)

	fmt.Println("========================================")
	fmt.Println("Mock Fulfillment Service Started")
	fmt.Println("========================================")
	fmt.Printf("Listening on: %s\n", listener.Addr().String())
	fmt.Println("")
	fmt.Println("To test with the CLI, run in another terminal:")
	fmt.Println("")
	fmt.Println("1. Login:")
	fmt.Printf("  ./fulfillment-cli login --plaintext http://127.0.0.1:%s\n", serverPort)
	fmt.Println("")
	fmt.Println("2. Test commands:")
	fmt.Println("  ./fulfillment-cli create computeinstance --template tpl-small-001 --name test-instance")
	fmt.Println("  ./fulfillment-cli describe computeinstance ci-mock-12345")
	fmt.Println("  ./fulfillment-cli get clusters --watch")
	fmt.Println("")
	fmt.Println("Press Ctrl+C to stop the server")
	fmt.Println("========================================")
	fmt.Println("")

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
