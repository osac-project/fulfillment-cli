/*
Copyright (c) 2025 Red Hat Inc.

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the
License. You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific
language governing permissions and limitations under the License.
*/

package reflection

import (
	"context"
	"strings"

	. "github.com/onsi/ginkgo/v2/dsl/core"
	. "github.com/onsi/ginkgo/v2/dsl/table"
	. "github.com/onsi/gomega"
	ffv1 "github.com/osac-project/fulfillment-common/api/fulfillment/v1"
	sharedv1 "github.com/osac-project/fulfillment-common/api/shared/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"

	"github.com/osac-project/fulfillment-cli/internal/testing"
)

var _ = Describe("Reflection helper", func() {
	var (
		ctx        context.Context
		server     *testing.Server
		connection *grpc.ClientConn
	)

	BeforeEach(func() {
		var err error

		// Create a context:
		ctx = context.Background()

		// Create the server:
		server = testing.NewServer()
		DeferCleanup(server.Stop)

		// Create the client connection:
		connection, err = grpc.NewClient(
			server.Address(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		Expect(err).ToNot(HaveOccurred())
		DeferCleanup(connection.Close)
	})

	Describe("Creation", func() {
		It("Can be created with all the mandatory parameters", func() {
			helper, err := NewHelper().
				SetLogger(logger).
				SetConnection(connection).
				AddPackage("fulfillment.v1", 1).
				Build()
			Expect(err).ToNot(HaveOccurred())
			Expect(helper).ToNot(BeNil())
		})

		It("Can be created with multiple packages", func() {
			helper, err := NewHelper().
				SetLogger(logger).
				SetConnection(connection).
				AddPackage("fulfillment.v1", 1).
				AddPackage("private.v1", 0).
				Build()
			Expect(err).ToNot(HaveOccurred())
			Expect(helper).ToNot(BeNil())
		})

		It("Can be created with multiple specified in one call", func() {
			helper, err := NewHelper().
				SetLogger(logger).
				SetConnection(connection).
				AddPackages(map[string]int{
					"private.v1":     0,
					"fulfillment.v1": 1,
				}).
				Build()
			Expect(err).ToNot(HaveOccurred())
			Expect(helper).ToNot(BeNil())
		})

		It("Can't be created without a logger", func() {
			helper, err := NewHelper().
				SetConnection(connection).
				AddPackage("fulfillment.v1", 1).
				Build()
			Expect(err).To(MatchError("logger is mandatory"))
			Expect(helper).To(BeNil())
		})

		It("Can't be created without a connection", func() {
			helper, err := NewHelper().
				SetLogger(logger).
				AddPackage("fulfillment.v1", 1).
				Build()
			Expect(err).To(MatchError("gRPC connection is mandatory"))
			Expect(helper).To(BeNil())
		})

		It("Can't be created without at least one package", func() {
			helper, err := NewHelper().
				SetLogger(logger).
				SetConnection(connection).
				Build()
			Expect(err).To(MatchError("at least one package is mandatory"))
			Expect(helper).To(BeNil())
		})
	})

	Describe("Behaviour", func() {
		var helper *Helper

		BeforeEach(func() {
			var err error
			helper, err = NewHelper().
				SetLogger(logger).
				SetConnection(connection).
				AddPackage("fulfillment.v1", 1).
				Build()
			Expect(err).ToNot(HaveOccurred())
		})

		It("Returns object types in singular", func() {
			Expect(helper.Singulars()).To(ConsistOf(
				"cluster",
				"clustertemplate",
				"computeinstance",
				"computeinstancetemplate",
				"host",
				"hostclass",
				"hostpool",
			))
		})

		It("Returns object types in plural", func() {
			Expect(helper.Plurals()).To(ConsistOf(
				"clusters",
				"clustertemplates",
				"computeinstances",
				"computeinstancetemplates",
				"hostclasses",
				"hostpools",
				"hosts",
			))
		})

		DescribeTable(
			"Lookup by object type",
			func(objectType string, expectedFullName string) {
				objectHelper := helper.Lookup(objectType)
				Expect(objectHelper).ToNot(BeNil())
				Expect(string(objectHelper.FullName())).To(Equal(expectedFullName))
			},
			Entry(
				"Cluster in singular",
				"cluster",
				"fulfillment.v1.Cluster",
			),
			Entry(
				"Cluster in plural",
				"clusters",
				"fulfillment.v1.Cluster",
			),
			Entry(
				"Cluster in singular upper case",
				"CLUSTER",
				"fulfillment.v1.Cluster",
			),
			Entry(
				"Host class in plural",
				"hostclasses",
				"fulfillment.v1.HostClass",
			),
			Entry(
				"Host in singular",
				"host",
				"fulfillment.v1.Host",
			),
			Entry(
				"Host in plural",
				"hosts",
				"fulfillment.v1.Host",
			),
			Entry(
				"Host pool in singular",
				"hostpool",
				"fulfillment.v1.HostPool",
			),
			Entry(
				"Host pool in plural",
				"hostpools",
				"fulfillment.v1.HostPool",
			),
		)

		DescribeTable(
			"Returns descriptor",
			func(objectType string, expectedFullName string) {
				objectHelper := helper.Lookup(objectType)
				Expect(objectHelper).ToNot(BeNil())
				objectDescriptor := objectHelper.Descriptor()
				Expect(objectDescriptor).ToNot(BeNil())
				Expect(string(objectDescriptor.FullName())).To(Equal(expectedFullName))
			},
			Entry(
				"Cluster",
				"cluster",
				"fulfillment.v1.Cluster",
			),
			Entry(
				"Cluster template",
				"clustertemplate",
				"fulfillment.v1.ClusterTemplate",
			),
			Entry(
				"Host class",
				"hostclass",
				"fulfillment.v1.HostClass",
			),
			Entry(
				"Compute instance template",
				"computeinstancetemplate",
				"fulfillment.v1.ComputeInstanceTemplate",
			),
			Entry(
				"Compute instance",
				"computeinstance",
				"fulfillment.v1.ComputeInstance",
			),
			Entry(
				"Host",
				"host",
				"fulfillment.v1.Host",
			),
			Entry(
				"Host pool",
				"hostpool",
				"fulfillment.v1.HostPool",
			),
		)

		DescribeTable(
			"Creates instance",
			func(objectType string, expectedInstance proto.Message) {
				objectHelper := helper.Lookup(objectType)
				Expect(objectHelper).ToNot(BeNil())
				actualInstance := objectHelper.Instance()
				Expect(proto.Equal(actualInstance, expectedInstance)).To(BeTrue())
			},
			Entry(
				"Cluster",
				"cluster",
				&ffv1.Cluster{},
			),
			Entry(
				"Cluster template",
				"clustertemplate",
				&ffv1.ClusterTemplate{},
			),
			Entry(
				"Host class",
				"hostclass",
				&ffv1.HostClass{},
			),
			Entry(
				"Host",
				"host",
				&ffv1.Host{},
			),
			Entry(
				"Host pool",
				"hostpool",
				&ffv1.HostPool{},
			),
		)

		It("Invokes get method", func() {
			// Register a clusters server that responds to the get request:
			ffv1.RegisterClustersServer(server.Registrar(), &testing.ClustersServerFuncs{
				GetFunc: func(ctx context.Context, request *ffv1.ClustersGetRequest,
				) (response *ffv1.ClustersGetResponse, err error) {
					defer GinkgoRecover()
					Expect(request.GetId()).To(Equal("123"))
					response = ffv1.ClustersGetResponse_builder{
						Object: ffv1.Cluster_builder{
							Id: "123",
							Status: ffv1.ClusterStatus_builder{
								State: ffv1.ClusterState_CLUSTER_STATE_READY,
							}.Build(),
						}.Build(),
					}.Build()
					return
				},
			})

			// Start the server:
			server.Start()

			// Use the helper to send the request, and verify the response:
			objectHelper := helper.Lookup("cluster")
			Expect(objectHelper).ToNot(BeNil())
			object, err := objectHelper.Get(ctx, "123")
			Expect(err).ToNot(HaveOccurred())
			Expect(proto.Equal(object, ffv1.Cluster_builder{
				Id: "123",
				Status: ffv1.ClusterStatus_builder{
					State: ffv1.ClusterState_CLUSTER_STATE_READY,
				}.Build(),
			}.Build())).To(BeTrue())
		})

		It("Invokes list method", func() {
			// Register a clusters server that responds to the list request:
			ffv1.RegisterClustersServer(server.Registrar(), &testing.ClustersServerFuncs{
				ListFunc: func(ctx context.Context, request *ffv1.ClustersListRequest,
				) (response *ffv1.ClustersListResponse, err error) {
					response = ffv1.ClustersListResponse_builder{
						Size:  proto.Int32(2),
						Total: proto.Int32(2),
						Items: []*ffv1.Cluster{
							ffv1.Cluster_builder{
								Id: "123",
							}.Build(),
							ffv1.Cluster_builder{
								Id: "456",
							}.Build(),
						},
					}.Build()
					return
				},
			})

			// Start the server:
			server.Start()

			// Use the helper to send the request, and verify the response:
			objectHelper := helper.Lookup("cluster")
			Expect(objectHelper).ToNot(BeNil())
			listResult, err := objectHelper.List(ctx, ListOptions{})
			Expect(err).ToNot(HaveOccurred())
			Expect(listResult.Items).To(HaveLen(2))
			Expect(listResult.Total).To(Equal(int32(2)))
			Expect(proto.Equal(
				listResult.Items[0],
				ffv1.Cluster_builder{
					Id: "123",
				}.Build(),
			)).To(BeTrue())
			Expect(proto.Equal(
				listResult.Items[1],
				ffv1.Cluster_builder{
					Id: "456",
				}.Build(),
			)).To(BeTrue())
		})

		It("Invokes create method", func() {
			// Register a clusters server that responds to the create request:
			ffv1.RegisterClustersServer(server.Registrar(), &testing.ClustersServerFuncs{
				CreateFunc: func(ctx context.Context, request *ffv1.ClustersCreateRequest,
				) (response *ffv1.ClustersCreateResponse, err error) {
					defer GinkgoRecover()
					Expect(proto.Equal(
						request.Object,
						ffv1.Cluster_builder{
							Spec: ffv1.ClusterSpec_builder{
								NodeSets: map[string]*ffv1.ClusterNodeSet{
									"xyz": ffv1.ClusterNodeSet_builder{
										HostClass: "acme_1tib",
										Size:      3,
									}.Build(),
								},
							}.Build(),
						}.Build(),
					)).To(BeTrue())
					response = ffv1.ClustersCreateResponse_builder{
						Object: ffv1.Cluster_builder{
							Id: "123",
							Spec: ffv1.ClusterSpec_builder{
								NodeSets: map[string]*ffv1.ClusterNodeSet{
									"xyz": ffv1.ClusterNodeSet_builder{
										HostClass: "acme_1tib",
										Size:      3,
									}.Build(),
								},
							}.Build(),
						}.Build(),
					}.Build()
					return
				},
			})

			// Start the server:
			server.Start()

			// Use the helper to send the request, and verify the response:
			objectHelper := helper.Lookup("cluster")
			Expect(objectHelper).ToNot(BeNil())
			object, err := objectHelper.Create(ctx, ffv1.Cluster_builder{
				Spec: ffv1.ClusterSpec_builder{
					NodeSets: map[string]*ffv1.ClusterNodeSet{
						"xyz": ffv1.ClusterNodeSet_builder{
							HostClass: "acme_1tib",
							Size:      3,
						}.Build(),
					},
				}.Build(),
			}.Build())
			Expect(err).ToNot(HaveOccurred())
			Expect(proto.Equal(
				object,
				ffv1.Cluster_builder{
					Id: "123",
					Spec: ffv1.ClusterSpec_builder{
						NodeSets: map[string]*ffv1.ClusterNodeSet{
							"xyz": ffv1.ClusterNodeSet_builder{
								HostClass: "acme_1tib",
								Size:      3,
							}.Build(),
						},
					}.Build(),
				}.Build(),
			)).To(BeTrue())
		})

		It("Invokes update method", func() {
			// Register a clusters server that responds to the update request:
			ffv1.RegisterClustersServer(server.Registrar(), &testing.ClustersServerFuncs{
				UpdateFunc: func(ctx context.Context, request *ffv1.ClustersUpdateRequest,
				) (response *ffv1.ClustersUpdateResponse, err error) {
					defer GinkgoRecover()
					Expect(proto.Equal(
						request.Object,
						ffv1.Cluster_builder{
							Id: "123",
							Spec: ffv1.ClusterSpec_builder{
								NodeSets: map[string]*ffv1.ClusterNodeSet{
									"xyz": ffv1.ClusterNodeSet_builder{
										Size: 3,
									}.Build(),
								},
							}.Build(),
						}.Build(),
					)).To(BeTrue())
					response = ffv1.ClustersUpdateResponse_builder{
						Object: ffv1.Cluster_builder{
							Id: "123",
							Spec: ffv1.ClusterSpec_builder{
								NodeSets: map[string]*ffv1.ClusterNodeSet{
									"xyz": ffv1.ClusterNodeSet_builder{
										HostClass: "acme_1tib",
										Size:      3,
									}.Build(),
								},
							}.Build(),
						}.Build(),
					}.Build()
					return
				},
			})

			// Start the server:
			server.Start()

			// Use the helper to send the request, and verify the response:
			objectHelper := helper.Lookup("cluster")
			Expect(objectHelper).ToNot(BeNil())
			object, err := objectHelper.Update(ctx, ffv1.Cluster_builder{
				Id: "123",
				Spec: ffv1.ClusterSpec_builder{
					NodeSets: map[string]*ffv1.ClusterNodeSet{
						"xyz": ffv1.ClusterNodeSet_builder{
							Size: 3,
						}.Build(),
					},
				}.Build(),
			}.Build())
			Expect(err).ToNot(HaveOccurred())
			Expect(proto.Equal(
				object,
				ffv1.Cluster_builder{
					Id: "123",
					Spec: ffv1.ClusterSpec_builder{
						NodeSets: map[string]*ffv1.ClusterNodeSet{
							"xyz": ffv1.ClusterNodeSet_builder{
								HostClass: "acme_1tib",
								Size:      3,
							}.Build(),
						},
					}.Build(),
				}.Build(),
			)).To(BeTrue())
		})

		It("Invokes delete method", func() {
			// Register a clusters server that responds to the delete request:
			ffv1.RegisterClustersServer(server.Registrar(), &testing.ClustersServerFuncs{
				DeleteFunc: func(ctx context.Context, request *ffv1.ClustersDeleteRequest,
				) (response *ffv1.ClustersDeleteResponse, err error) {
					defer GinkgoRecover()
					response = ffv1.ClustersDeleteResponse_builder{}.Build()
					return
				},
			})

			// Start the server:
			server.Start()

			// Use the helper to send the request, and verify the response:
			objectHelper := helper.Lookup("cluster")
			Expect(objectHelper).ToNot(BeNil())
			err := objectHelper.Delete(ctx, "123")
			Expect(err).ToNot(HaveOccurred())
		})

		It("Invokes hosts get method", func() {
			// Register a hosts server that responds to the get request:
			ffv1.RegisterHostsServer(server.Registrar(), &testing.HostsServerFuncs{
				GetFunc: func(ctx context.Context, request *ffv1.HostsGetRequest,
				) (response *ffv1.HostsGetResponse, err error) {
					defer GinkgoRecover()
					Expect(request.GetId()).To(Equal("host-123"))
					response = ffv1.HostsGetResponse_builder{
						Object: ffv1.Host_builder{
							Id: "host-123",
							Status: ffv1.HostStatus_builder{
								PowerState: ffv1.HostPowerState_HOST_POWER_STATE_ON,
							}.Build(),
						}.Build(),
					}.Build()
					return
				},
			})

			// Start the server:
			server.Start()

			// Use the helper to send the request, and verify the response:
			objectHelper := helper.Lookup("host")
			Expect(objectHelper).ToNot(BeNil())
			object, err := objectHelper.Get(ctx, "host-123")
			Expect(err).ToNot(HaveOccurred())
			Expect(proto.Equal(object, ffv1.Host_builder{
				Id: "host-123",
				Status: ffv1.HostStatus_builder{
					PowerState: ffv1.HostPowerState_HOST_POWER_STATE_ON,
				}.Build(),
			}.Build())).To(BeTrue())
		})

		It("Invokes hosts delete method", func() {
			// Register a hosts server that responds to the delete request:
			ffv1.RegisterHostsServer(server.Registrar(), &testing.HostsServerFuncs{
				DeleteFunc: func(ctx context.Context, request *ffv1.HostsDeleteRequest,
				) (response *ffv1.HostsDeleteResponse, err error) {
					defer GinkgoRecover()
					Expect(request.GetId()).To(Equal("host-123"))
					response = ffv1.HostsDeleteResponse_builder{}.Build()
					return
				},
			})

			// Start the server:
			server.Start()

			// Use the helper to send the request, and verify the response:
			objectHelper := helper.Lookup("host")
			Expect(objectHelper).ToNot(BeNil())
			err := objectHelper.Delete(ctx, "host-123")
			Expect(err).ToNot(HaveOccurred())
		})

		It("Invokes host pools get method", func() {
			// Register a host pools server that responds to the get request:
			ffv1.RegisterHostPoolsServer(server.Registrar(), &testing.HostPoolsServerFuncs{
				GetFunc: func(ctx context.Context, request *ffv1.HostPoolsGetRequest,
				) (response *ffv1.HostPoolsGetResponse, err error) {
					defer GinkgoRecover()
					Expect(request.GetId()).To(Equal("pool-123"))
					response = ffv1.HostPoolsGetResponse_builder{
						Object: ffv1.HostPool_builder{
							Id: "pool-123",
							Status: ffv1.HostPoolStatus_builder{
								State: ffv1.HostPoolState_HOST_POOL_STATE_READY,
							}.Build(),
						}.Build(),
					}.Build()
					return
				},
			})

			// Start the server:
			server.Start()

			// Use the helper to send the request, and verify the response:
			objectHelper := helper.Lookup("hostpool")
			Expect(objectHelper).ToNot(BeNil())
			object, err := objectHelper.Get(ctx, "pool-123")
			Expect(err).ToNot(HaveOccurred())
			Expect(proto.Equal(object, ffv1.HostPool_builder{
				Id: "pool-123",
				Status: ffv1.HostPoolStatus_builder{
					State: ffv1.HostPoolState_HOST_POOL_STATE_READY,
				}.Build(),
			}.Build())).To(BeTrue())
		})

		It("Invokes host pools delete method", func() {
			// Register a host pools server that responds to the delete request:
			ffv1.RegisterHostPoolsServer(server.Registrar(), &testing.HostPoolsServerFuncs{
				DeleteFunc: func(ctx context.Context, request *ffv1.HostPoolsDeleteRequest,
				) (response *ffv1.HostPoolsDeleteResponse, err error) {
					defer GinkgoRecover()
					Expect(request.GetId()).To(Equal("pool-123"))
					response = ffv1.HostPoolsDeleteResponse_builder{}.Build()
					return
				},
			})

			// Start the server:
			server.Start()

			// Use the helper to send the request, and verify the response:
			objectHelper := helper.Lookup("hostpool")
			Expect(objectHelper).ToNot(BeNil())
			err := objectHelper.Delete(ctx, "pool-123")
			Expect(err).ToNot(HaveOccurred())
		})

		It("Returns metadata from get method", func() {
			// Register a clusters server that responds to the get request with metadata:
			ffv1.RegisterClustersServer(server.Registrar(), &testing.ClustersServerFuncs{
				GetFunc: func(ctx context.Context, request *ffv1.ClustersGetRequest,
				) (response *ffv1.ClustersGetResponse, err error) {
					defer GinkgoRecover()
					Expect(request.GetId()).To(Equal("123"))
					response = ffv1.ClustersGetResponse_builder{
						Object: ffv1.Cluster_builder{
							Id: "123",
							Metadata: sharedv1.Metadata_builder{
								Name: "my-cluster",
							}.Build(),
						}.Build(),
					}.Build()
					return
				},
			})

			// Start the server:
			server.Start()

			// Use the helper to send the request, and verify the metadata is returned:
			objectHelper := helper.Lookup("cluster")
			Expect(objectHelper).ToNot(BeNil())
			object, err := objectHelper.Get(ctx, "123")
			Expect(err).ToNot(HaveOccurred())
			Expect(object).ToNot(BeNil())
			metadata := objectHelper.GetMetadata(object)
			Expect(metadata).ToNot(BeNil())
			Expect(metadata.GetName()).To(Equal("my-cluster"))
		})

		It("Sorts types according to package order", func() {
			// Create a helper with multiple packages, where 'private.v1' has a lower order (0) than
			// 'fulfillment.v1' (1), so 'private.v1' types should appear first:
			multiPackageHelper, err := NewHelper().
				SetLogger(logger).
				SetConnection(connection).
				AddPackage("private.v1", 0).
				AddPackage("fulfillment.v1", 1).
				Build()
			Expect(err).ToNot(HaveOccurred())

			// Get all the type names:
			names := multiPackageHelper.Names()

			// Verify that 'private.v1' types come before 'fulfillment.v1' types:
			var lastPrivateIndex int = -1
			var firstFulfillmentIndex int = -1
			for i, name := range names {
				if strings.HasPrefix(name, "private.v1.") {
					lastPrivateIndex = i
				}
				if strings.HasPrefix(name, "fulfillment.v1.") && firstFulfillmentIndex == -1 {
					firstFulfillmentIndex = i
				}
			}

			// If both package types exist, verify that all 'private.v1' types come before
			// 'fulfillment.v1' types:
			if lastPrivateIndex >= 0 && firstFulfillmentIndex >= 0 {
				Expect(lastPrivateIndex).To(
					BeNumerically("<", firstFulfillmentIndex),
					"All 'private.v1' types should come before 'fulfillment.v1' types",
				)
			}

			// Verify that within each package, types are sorted alphabetically:
			privateTypes := []string{}
			fulfillmentTypes := []string{}
			for _, name := range names {
				if strings.HasPrefix(name, "private.v1.") {
					privateTypes = append(privateTypes, name)
				}
				if strings.HasPrefix(name, "fulfillment.v1.") {
					fulfillmentTypes = append(fulfillmentTypes, name)
				}
			}

			// Check that 'private.v1' types are sorted alphabetically:
			if len(privateTypes) > 1 {
				for i := 1; i < len(privateTypes); i++ {
					Expect(privateTypes[i-1] < privateTypes[i]).To(
						BeTrue(),
						"Types within 'private.v1' should be sorted alphabetically, '%s' "+
							"should come before '%s'",
						privateTypes[i-1], privateTypes[i],
					)
				}
			}

			// Check that 'fulfillment.v1' types are sorted alphabetically:
			if len(fulfillmentTypes) > 1 {
				for i := 1; i < len(fulfillmentTypes); i++ {
					Expect(fulfillmentTypes[i-1] < fulfillmentTypes[i]).To(
						BeTrue(),
						"Types within 'fulfillment.v1' should be sorted alphabetically, '%s' "+
							"should come before '%s'",
						fulfillmentTypes[i-1], fulfillmentTypes[i],
					)
				}
			}
		})

		It("Sorts types according to package order when adding packages", func() {
			// Create a helper using AddPackages method with reversed order:
			multiPackageHelper, err := NewHelper().
				SetLogger(logger).
				SetConnection(connection).
				AddPackages(map[string]int{
					"fulfillment.v1": 2,
					"private.v1":     1,
				}).
				Build()
			Expect(err).ToNot(HaveOccurred())

			// Get all the type names:
			names := multiPackageHelper.Names()

			// Verify that 'private.v1' types come before 'fulfillment.v1' types, since 'private.v1' has
			// order 1, 'fulfillment.v1' has order 2:
			var lastPrivateIndex int = -1
			var firstFulfillmentIndex int = -1
			for i, name := range names {
				if strings.HasPrefix(name, "private.v1.") {
					lastPrivateIndex = i
				}
				if strings.HasPrefix(name, "fulfillment.v1.") && firstFulfillmentIndex == -1 {
					firstFulfillmentIndex = i
				}
			}

			// If both package types exist, verify that all 'private.v1' types come before 'fulfillment.v1'
			// types:
			if lastPrivateIndex >= 0 && firstFulfillmentIndex >= 0 {
				Expect(lastPrivateIndex).To(
					BeNumerically("<", firstFulfillmentIndex),
					"All 'private.v1' types should come before 'fulfillment.v1' types when "+
						"'private.v1' has lower order",
				)
			}
		})
	})
})
