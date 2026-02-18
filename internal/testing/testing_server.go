/*
Copyright (c) 2025 Red Hat Inc.

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the
License. You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific
language governing permissions and limitations under the License.
*/

package testing

import (
	"context"
	"net"
	"strings"

	. "github.com/onsi/ginkgo/v2/dsl/core"
	. "github.com/onsi/gomega"
	eventsv1 "github.com/osac-project/fulfillment-common/api/events/v1"
	ffv1 "github.com/osac-project/fulfillment-common/api/fulfillment/v1"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc"
)

// Server is a gRPC server used only for tests.
type Server struct {
	listener net.Listener
	server   *grpc.Server
}

// NewServer creates a new gRPC server that listens in a randomly selected port in the local host.
func NewServer() *Server {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	Expect(err).ToNot(HaveOccurred())
	server := grpc.NewServer()
	return &Server{
		listener: listener,
		server:   server,
	}
}

// Adress returns the address where the server is listening.
func (s *Server) Address() string {
	return s.listener.Addr().String()
}

// Registrar returns the registrar that can be used to register server implementations.
func (s *Server) Registrar() grpc.ServiceRegistrar {
	return s.server
}

// Start starts the server. This needs to be done after registering all server implementations, and before trying to
// call any of them.
func (s *Server) Start() {
	go func() {
		defer GinkgoRecover()
		err := s.server.Serve(s.listener)
		Expect(err).ToNot(HaveOccurred())
	}()
}

// Stop stops the server, closing all connections and releasing all the resources it was using.
func (s *Server) Stop() {
	s.server.Stop()
}

// Make sure that we implement the interface.
var _ ffv1.ClustersServer = (*ClustersServerFuncs)(nil)

// ClustersServerFuncs is an implementation of the clusters server that uses configurable functions to implement the
// methods.
type ClustersServerFuncs struct {
	ffv1.UnimplementedClustersServer

	CreateFunc               func(context.Context, *ffv1.ClustersCreateRequest) (*ffv1.ClustersCreateResponse, error)
	DeleteFunc               func(context.Context, *ffv1.ClustersDeleteRequest) (*ffv1.ClustersDeleteResponse, error)
	GetFunc                  func(context.Context, *ffv1.ClustersGetRequest) (*ffv1.ClustersGetResponse, error)
	ListFunc                 func(context.Context, *ffv1.ClustersListRequest) (*ffv1.ClustersListResponse, error)
	GetKubeconfigFunc        func(context.Context, *ffv1.ClustersGetKubeconfigRequest) (*ffv1.ClustersGetKubeconfigResponse, error)
	GetKubeconfigViaHttpFunc func(context.Context, *ffv1.ClustersGetKubeconfigViaHttpRequest) (*httpbody.HttpBody, error)
	UpdateFunc               func(context.Context, *ffv1.ClustersUpdateRequest) (*ffv1.ClustersUpdateResponse, error)
}

func (s *ClustersServerFuncs) Create(ctx context.Context,
	request *ffv1.ClustersCreateRequest) (response *ffv1.ClustersCreateResponse, err error) {
	response, err = s.CreateFunc(ctx, request)
	return
}

func (s *ClustersServerFuncs) Delete(ctx context.Context,
	request *ffv1.ClustersDeleteRequest) (response *ffv1.ClustersDeleteResponse, err error) {
	response, err = s.DeleteFunc(ctx, request)
	return
}

func (s *ClustersServerFuncs) Get(ctx context.Context,
	request *ffv1.ClustersGetRequest) (response *ffv1.ClustersGetResponse, err error) {
	response, err = s.GetFunc(ctx, request)
	return
}

func (s *ClustersServerFuncs) GetKubeconfig(ctx context.Context,
	request *ffv1.ClustersGetKubeconfigRequest) (response *ffv1.ClustersGetKubeconfigResponse, err error) {
	response, err = s.GetKubeconfigFunc(ctx, request)
	return
}

func (s *ClustersServerFuncs) GetKubeconfigViaHttp(ctx context.Context,
	request *ffv1.ClustersGetKubeconfigViaHttpRequest) (response *httpbody.HttpBody, err error) {
	response, err = s.GetKubeconfigViaHttpFunc(ctx, request)
	return
}

func (s *ClustersServerFuncs) List(ctx context.Context,
	request *ffv1.ClustersListRequest) (response *ffv1.ClustersListResponse, err error) {
	response, err = s.ListFunc(ctx, request)
	return
}

func (s *ClustersServerFuncs) Update(ctx context.Context,
	request *ffv1.ClustersUpdateRequest) (response *ffv1.ClustersUpdateResponse, err error) {
	response, err = s.UpdateFunc(ctx, request)
	return
}

// Make sure that we implement the interface.
var _ ffv1.HostsServer = (*HostsServerFuncs)(nil)

// HostsServerFuncs is an implementation of the hosts server that uses configurable functions to implement the
// methods.
type HostsServerFuncs struct {
	ffv1.UnimplementedHostsServer

	CreateFunc func(context.Context, *ffv1.HostsCreateRequest) (*ffv1.HostsCreateResponse, error)
	DeleteFunc func(context.Context, *ffv1.HostsDeleteRequest) (*ffv1.HostsDeleteResponse, error)
	GetFunc    func(context.Context, *ffv1.HostsGetRequest) (*ffv1.HostsGetResponse, error)
	ListFunc   func(context.Context, *ffv1.HostsListRequest) (*ffv1.HostsListResponse, error)
	UpdateFunc func(context.Context, *ffv1.HostsUpdateRequest) (*ffv1.HostsUpdateResponse, error)
}

func (s *HostsServerFuncs) Create(ctx context.Context,
	request *ffv1.HostsCreateRequest) (response *ffv1.HostsCreateResponse, err error) {
	response, err = s.CreateFunc(ctx, request)
	return
}

func (s *HostsServerFuncs) Delete(ctx context.Context,
	request *ffv1.HostsDeleteRequest) (response *ffv1.HostsDeleteResponse, err error) {
	response, err = s.DeleteFunc(ctx, request)
	return
}

func (s *HostsServerFuncs) Get(ctx context.Context,
	request *ffv1.HostsGetRequest) (response *ffv1.HostsGetResponse, err error) {
	response, err = s.GetFunc(ctx, request)
	return
}

func (s *HostsServerFuncs) List(ctx context.Context,
	request *ffv1.HostsListRequest) (response *ffv1.HostsListResponse, err error) {
	response, err = s.ListFunc(ctx, request)
	return
}

func (s *HostsServerFuncs) Update(ctx context.Context,
	request *ffv1.HostsUpdateRequest) (response *ffv1.HostsUpdateResponse, err error) {
	response, err = s.UpdateFunc(ctx, request)
	return
}

// Make sure that we implement the interface.
var _ ffv1.HostPoolsServer = (*HostPoolsServerFuncs)(nil)

// HostPoolsServerFuncs is an implementation of the host pools server that uses configurable functions to implement the
// methods.
type HostPoolsServerFuncs struct {
	ffv1.UnimplementedHostPoolsServer

	CreateFunc func(context.Context, *ffv1.HostPoolsCreateRequest) (*ffv1.HostPoolsCreateResponse, error)
	DeleteFunc func(context.Context, *ffv1.HostPoolsDeleteRequest) (*ffv1.HostPoolsDeleteResponse, error)
	GetFunc    func(context.Context, *ffv1.HostPoolsGetRequest) (*ffv1.HostPoolsGetResponse, error)
	ListFunc   func(context.Context, *ffv1.HostPoolsListRequest) (*ffv1.HostPoolsListResponse, error)
	UpdateFunc func(context.Context, *ffv1.HostPoolsUpdateRequest) (*ffv1.HostPoolsUpdateResponse, error)
}

func (s *HostPoolsServerFuncs) Create(ctx context.Context,
	request *ffv1.HostPoolsCreateRequest) (response *ffv1.HostPoolsCreateResponse, err error) {
	response, err = s.CreateFunc(ctx, request)
	return
}

func (s *HostPoolsServerFuncs) Delete(ctx context.Context,
	request *ffv1.HostPoolsDeleteRequest) (response *ffv1.HostPoolsDeleteResponse, err error) {
	response, err = s.DeleteFunc(ctx, request)
	return
}

func (s *HostPoolsServerFuncs) Get(ctx context.Context,
	request *ffv1.HostPoolsGetRequest) (response *ffv1.HostPoolsGetResponse, err error) {
	response, err = s.GetFunc(ctx, request)
	return
}

func (s *HostPoolsServerFuncs) List(ctx context.Context,
	request *ffv1.HostPoolsListRequest) (response *ffv1.HostPoolsListResponse, err error) {
	response, err = s.ListFunc(ctx, request)
	return
}

func (s *HostPoolsServerFuncs) Update(ctx context.Context,
	request *ffv1.HostPoolsUpdateRequest) (response *ffv1.HostPoolsUpdateResponse, err error) {
	response, err = s.UpdateFunc(ctx, request)
	return
}

// Make sure that we implement the interface.
var _ ffv1.ComputeInstancesServer = (*ComputeInstancesServerFuncs)(nil)

// ComputeInstancesServerFuncs is an implementation of the compute instances server that uses configurable functions to implement the
// methods.
type ComputeInstancesServerFuncs struct {
	ffv1.UnimplementedComputeInstancesServer

	CreateFunc func(context.Context, *ffv1.ComputeInstancesCreateRequest) (*ffv1.ComputeInstancesCreateResponse, error)
	DeleteFunc func(context.Context, *ffv1.ComputeInstancesDeleteRequest) (*ffv1.ComputeInstancesDeleteResponse, error)
	GetFunc    func(context.Context, *ffv1.ComputeInstancesGetRequest) (*ffv1.ComputeInstancesGetResponse, error)
	ListFunc   func(context.Context, *ffv1.ComputeInstancesListRequest) (*ffv1.ComputeInstancesListResponse, error)
	UpdateFunc func(context.Context, *ffv1.ComputeInstancesUpdateRequest) (*ffv1.ComputeInstancesUpdateResponse, error)
}

func (s *ComputeInstancesServerFuncs) Create(ctx context.Context,
	request *ffv1.ComputeInstancesCreateRequest) (response *ffv1.ComputeInstancesCreateResponse, err error) {
	response, err = s.CreateFunc(ctx, request)
	return
}

func (s *ComputeInstancesServerFuncs) Delete(ctx context.Context,
	request *ffv1.ComputeInstancesDeleteRequest) (response *ffv1.ComputeInstancesDeleteResponse, err error) {
	response, err = s.DeleteFunc(ctx, request)
	return
}

func (s *ComputeInstancesServerFuncs) Get(ctx context.Context,
	request *ffv1.ComputeInstancesGetRequest) (response *ffv1.ComputeInstancesGetResponse, err error) {
	response, err = s.GetFunc(ctx, request)
	return
}

func (s *ComputeInstancesServerFuncs) List(ctx context.Context,
	request *ffv1.ComputeInstancesListRequest) (response *ffv1.ComputeInstancesListResponse, err error) {
	response, err = s.ListFunc(ctx, request)
	return
}

func (s *ComputeInstancesServerFuncs) Update(ctx context.Context,
	request *ffv1.ComputeInstancesUpdateRequest) (response *ffv1.ComputeInstancesUpdateResponse, err error) {
	response, err = s.UpdateFunc(ctx, request)
	return
}

// Make sure that we implement the interface.
var _ ffv1.ComputeInstanceTemplatesServer = (*ComputeInstanceTemplatesServerFuncs)(nil)

// ComputeInstanceTemplatesServerFuncs is an implementation of the compute instance templates server that uses configurable functions to implement the
// methods.
type ComputeInstanceTemplatesServerFuncs struct {
	ffv1.UnimplementedComputeInstanceTemplatesServer

	CreateFunc func(context.Context, *ffv1.ComputeInstanceTemplatesCreateRequest) (*ffv1.ComputeInstanceTemplatesCreateResponse, error)
	DeleteFunc func(context.Context, *ffv1.ComputeInstanceTemplatesDeleteRequest) (*ffv1.ComputeInstanceTemplatesDeleteResponse, error)
	GetFunc    func(context.Context, *ffv1.ComputeInstanceTemplatesGetRequest) (*ffv1.ComputeInstanceTemplatesGetResponse, error)
	ListFunc   func(context.Context, *ffv1.ComputeInstanceTemplatesListRequest) (*ffv1.ComputeInstanceTemplatesListResponse, error)
	UpdateFunc func(context.Context, *ffv1.ComputeInstanceTemplatesUpdateRequest) (*ffv1.ComputeInstanceTemplatesUpdateResponse, error)
}

func (s *ComputeInstanceTemplatesServerFuncs) Create(ctx context.Context,
	request *ffv1.ComputeInstanceTemplatesCreateRequest) (response *ffv1.ComputeInstanceTemplatesCreateResponse, err error) {
	response, err = s.CreateFunc(ctx, request)
	return
}

func (s *ComputeInstanceTemplatesServerFuncs) Delete(ctx context.Context,
	request *ffv1.ComputeInstanceTemplatesDeleteRequest) (response *ffv1.ComputeInstanceTemplatesDeleteResponse, err error) {
	response, err = s.DeleteFunc(ctx, request)
	return
}

func (s *ComputeInstanceTemplatesServerFuncs) Get(ctx context.Context,
	request *ffv1.ComputeInstanceTemplatesGetRequest) (response *ffv1.ComputeInstanceTemplatesGetResponse, err error) {
	response, err = s.GetFunc(ctx, request)
	return
}

func (s *ComputeInstanceTemplatesServerFuncs) List(ctx context.Context,
	request *ffv1.ComputeInstanceTemplatesListRequest) (response *ffv1.ComputeInstanceTemplatesListResponse, err error) {
	response, err = s.ListFunc(ctx, request)
	return
}

func (s *ComputeInstanceTemplatesServerFuncs) Update(ctx context.Context,
	request *ffv1.ComputeInstanceTemplatesUpdateRequest) (response *ffv1.ComputeInstanceTemplatesUpdateResponse, err error) {
	response, err = s.UpdateFunc(ctx, request)
	return
}

// Make sure that we implement the interface.
var _ eventsv1.EventsServer = (*EventsServerFuncs)(nil)

// EventsServerFuncs is an implementation of the events server that uses configurable functions to implement the
// methods.
type EventsServerFuncs struct {
	eventsv1.UnimplementedEventsServer

	WatchFunc func(*eventsv1.EventsWatchRequest, eventsv1.Events_WatchServer) error
}

func (s *EventsServerFuncs) Watch(request *eventsv1.EventsWatchRequest, stream eventsv1.Events_WatchServer) error {
	return s.WatchFunc(request, stream)
}

// Helper function to extract object ID from event
func GetEventObjectID(event *eventsv1.Event) string {
	switch payload := event.Payload.(type) {
	case *eventsv1.Event_Cluster:
		if payload.Cluster != nil {
			return payload.Cluster.Id
		}
	case *eventsv1.Event_ClusterTemplate:
		if payload.ClusterTemplate != nil {
			return payload.ClusterTemplate.Id
		}
	}
	return ""
}

// Helper function to check if an event matches the filter
func MatchesFilter(event *eventsv1.Event, filter string) bool {
	// Empty filter - send all events
	if filter == "" {
		return true
	}

	// Determine what type the filter is for
	// Check for cluster_template first (more specific) to avoid substring issues
	isClusterTemplateFilter := strings.Contains(filter, "event.cluster_template")
	isClusterFilter := strings.Contains(filter, "event.cluster") && !isClusterTemplateFilter

	// Check if the event type matches the filter type
	switch event.Payload.(type) {
	case *eventsv1.Event_Cluster:
		if !isClusterFilter {
			return false
		}
		// If filter is just a type check, send all cluster events
		if filter == "has(event.cluster)" {
			return true
		}
	case *eventsv1.Event_ClusterTemplate:
		if !isClusterTemplateFilter {
			return false
		}
		// If filter is just a type check, send all cluster template events
		if filter == "has(event.cluster_template)" {
			return true
		}
	default:
		return false
	}

	// Extract ID and name from the event
	var id, name string
	switch payload := event.Payload.(type) {
	case *eventsv1.Event_Cluster:
		if payload.Cluster != nil {
			id = payload.Cluster.Id
			if payload.Cluster.Metadata != nil {
				name = payload.Cluster.Metadata.Name
			}
		}
	case *eventsv1.Event_ClusterTemplate:
		if payload.ClusterTemplate != nil {
			id = payload.ClusterTemplate.Id
			if payload.ClusterTemplate.Metadata != nil {
				name = payload.ClusterTemplate.Metadata.Name
			}
		}
	}

	// Check if filter contains the specific ID or name
	return strings.Contains(filter, id) || strings.Contains(filter, name)
}

// Helper function to send an event if it matches the filter
func SendEventIfMatches(event *eventsv1.Event, filter string, stream eventsv1.Events_WatchServer) error {
	if MatchesFilter(event, filter) {
		return stream.Send(&eventsv1.EventsWatchResponse{Event: event})
	}
	return nil
}
