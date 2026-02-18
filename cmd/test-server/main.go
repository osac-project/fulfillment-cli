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
	"strings"

	eventsv1 "github.com/osac-project/fulfillment-common/api/events/v1"
	ffv1 "github.com/osac-project/fulfillment-common/api/fulfillment/v1"
	metadatav1 "github.com/osac-project/fulfillment-common/api/metadata/v1"
	sharedv1 "github.com/osac-project/fulfillment-common/api/shared/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/osac-project/fulfillment-cli/internal/testing"
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

// Simple mock compute instances server for testing
type computeInstancesServer struct {
	ffv1.UnimplementedComputeInstancesServer
}

func (s *computeInstancesServer) Create(ctx context.Context, request *ffv1.ComputeInstancesCreateRequest) (*ffv1.ComputeInstancesCreateResponse, error) {
	instance := request.GetObject()

	// Set mock ID and state if not already set
	if instance.Id == "" {
		instance.Id = "ci-mock-12345"
	}
	if instance.Status == nil {
		instance.Status = &ffv1.ComputeInstanceStatus{
			State: ffv1.ComputeInstanceState_COMPUTE_INSTANCE_STATE_STARTING,
		}
	}

	log.Printf("Created compute instance: %s (name: %s, template: %s)",
		instance.Id,
		instance.GetMetadata().GetName(),
		instance.GetSpec().GetTemplate())

	return &ffv1.ComputeInstancesCreateResponse{Object: instance}, nil
}

func (s *computeInstancesServer) Get(ctx context.Context, request *ffv1.ComputeInstancesGetRequest) (*ffv1.ComputeInstancesGetResponse, error) {
	// Return a mock instance
	instance := &ffv1.ComputeInstance{
		Id: request.Id,
		Metadata: &sharedv1.Metadata{
			Name: "mock-instance",
		},
		Spec: &ffv1.ComputeInstanceSpec{
			Template: "small-instance",
		},
		Status: &ffv1.ComputeInstanceStatus{
			State:     ffv1.ComputeInstanceState_COMPUTE_INSTANCE_STATE_RUNNING,
			IpAddress: "192.168.1.100",
		},
	}

	log.Printf("Retrieved compute instance: %s", request.Id)
	return &ffv1.ComputeInstancesGetResponse{Object: instance}, nil
}

func (s *computeInstancesServer) List(ctx context.Context, request *ffv1.ComputeInstancesListRequest) (*ffv1.ComputeInstancesListResponse, error) {
	// Return a mock list with one instance
	instance := &ffv1.ComputeInstance{
		Id: "ci-mock-12345",
		Metadata: &sharedv1.Metadata{
			Name: "mock-instance",
		},
		Spec: &ffv1.ComputeInstanceSpec{
			Template: "small-instance",
		},
		Status: &ffv1.ComputeInstanceStatus{
			State:     ffv1.ComputeInstanceState_COMPUTE_INSTANCE_STATE_RUNNING,
			IpAddress: "192.168.1.100",
		},
	}

	size := int32(1)
	total := int32(1)
	log.Printf("Listed compute instances")
	return &ffv1.ComputeInstancesListResponse{
		Items: []*ffv1.ComputeInstance{instance},
		Size:  &size,
		Total: &total,
	}, nil
}

// Simple mock compute instance templates server
type computeInstanceTemplatesServer struct {
	ffv1.UnimplementedComputeInstanceTemplatesServer
}

func (s *computeInstanceTemplatesServer) Get(ctx context.Context, request *ffv1.ComputeInstanceTemplatesGetRequest) (*ffv1.ComputeInstanceTemplatesGetResponse, error) {
	// Return a mock template
	template := &ffv1.ComputeInstanceTemplate{
		Id: request.Id,
		Metadata: &sharedv1.Metadata{
			Name: request.Id,
		},
	}

	log.Printf("Retrieved compute instance template: %s", request.Id)
	return &ffv1.ComputeInstanceTemplatesGetResponse{Object: template}, nil
}

func (s *computeInstanceTemplatesServer) List(ctx context.Context, request *ffv1.ComputeInstanceTemplatesListRequest) (*ffv1.ComputeInstanceTemplatesListResponse, error) {
	// All available templates
	allTemplates := []*ffv1.ComputeInstanceTemplate{
		{
			Id: "tpl-small-001",
			Metadata: &sharedv1.Metadata{
				Name: "small-instance",
			},
		},
		{
			Id: "tpl-large-001",
			Metadata: &sharedv1.Metadata{
				Name: "large-instance",
			},
		},
	}

	// Apply filter if provided (simple string matching for mock purposes)
	filter := request.GetFilter()
	var templates []*ffv1.ComputeInstanceTemplate

	if filter != "" {
		// Simple filter: check if filter contains the template ID or name
		// This is a mock implementation - real server would parse CEL expressions
		for _, tmpl := range allTemplates {
			// Check if filter mentions this template's ID or name
			if strings.Contains(filter, tmpl.Id) || strings.Contains(filter, tmpl.GetMetadata().GetName()) {
				templates = append(templates, tmpl)
			}
		}
	} else {
		templates = allTemplates
	}

	size := int32(len(templates))
	total := int32(len(templates))
	log.Printf("Listed compute instance templates (filter: %q, matches: %d)", filter, len(templates))
	return &ffv1.ComputeInstanceTemplatesListResponse{
		Items: templates,
		Size:  &size,
		Total: &total,
	}, nil
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

	// Load scenario from file
	scenario, err := testing.LoadScenarioFromFile(*scenarioFile)
	if err != nil {
		log.Fatalf("Failed to load scenario from %s: %v", *scenarioFile, err)
	}
	log.Printf("Loaded scenario: %s - %s", scenario.Name, scenario.Description)

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
	ffv1.RegisterComputeInstancesServer(grpcServer, &computeInstancesServer{})
	ffv1.RegisterComputeInstanceTemplatesServer(grpcServer, &computeInstanceTemplatesServer{})
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
