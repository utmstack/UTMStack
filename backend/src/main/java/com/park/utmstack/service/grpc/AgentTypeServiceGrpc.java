package com.park.utmstack.service.grpc;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.55.1)",
    comments = "Source: agent.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class AgentTypeServiceGrpc {

  private AgentTypeServiceGrpc() {}

  public static final String SERVICE_NAME = "agent.AgentTypeService";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.ListRequest,
      com.park.utmstack.service.grpc.ListAgentsTypeResponse> getListAgentTypesMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ListAgentTypes",
      requestType = com.park.utmstack.service.grpc.ListRequest.class,
      responseType = com.park.utmstack.service.grpc.ListAgentsTypeResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.ListRequest,
      com.park.utmstack.service.grpc.ListAgentsTypeResponse> getListAgentTypesMethod() {
    io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.ListRequest, com.park.utmstack.service.grpc.ListAgentsTypeResponse> getListAgentTypesMethod;
    if ((getListAgentTypesMethod = AgentTypeServiceGrpc.getListAgentTypesMethod) == null) {
      synchronized (AgentTypeServiceGrpc.class) {
        if ((getListAgentTypesMethod = AgentTypeServiceGrpc.getListAgentTypesMethod) == null) {
          AgentTypeServiceGrpc.getListAgentTypesMethod = getListAgentTypesMethod =
              io.grpc.MethodDescriptor.<com.park.utmstack.service.grpc.ListRequest, com.park.utmstack.service.grpc.ListAgentsTypeResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ListAgentTypes"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.park.utmstack.service.grpc.ListRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.park.utmstack.service.grpc.ListAgentsTypeResponse.getDefaultInstance()))
              .setSchemaDescriptor(new AgentTypeServiceMethodDescriptorSupplier("ListAgentTypes"))
              .build();
        }
      }
    }
    return getListAgentTypesMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static AgentTypeServiceStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<AgentTypeServiceStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<AgentTypeServiceStub>() {
        @java.lang.Override
        public AgentTypeServiceStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new AgentTypeServiceStub(channel, callOptions);
        }
      };
    return AgentTypeServiceStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static AgentTypeServiceBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<AgentTypeServiceBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<AgentTypeServiceBlockingStub>() {
        @java.lang.Override
        public AgentTypeServiceBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new AgentTypeServiceBlockingStub(channel, callOptions);
        }
      };
    return AgentTypeServiceBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static AgentTypeServiceFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<AgentTypeServiceFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<AgentTypeServiceFutureStub>() {
        @java.lang.Override
        public AgentTypeServiceFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new AgentTypeServiceFutureStub(channel, callOptions);
        }
      };
    return AgentTypeServiceFutureStub.newStub(factory, channel);
  }

  /**
   */
  public interface AsyncService {

    /**
     */
    default void listAgentTypes(com.park.utmstack.service.grpc.ListRequest request,
        io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.ListAgentsTypeResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getListAgentTypesMethod(), responseObserver);
    }
  }

  /**
   * Base class for the server implementation of the service AgentTypeService.
   */
  public static abstract class AgentTypeServiceImplBase
      implements io.grpc.BindableService, AsyncService {

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return AgentTypeServiceGrpc.bindService(this);
    }
  }

  /**
   * A stub to allow clients to do asynchronous rpc calls to service AgentTypeService.
   */
  public static final class AgentTypeServiceStub
      extends io.grpc.stub.AbstractAsyncStub<AgentTypeServiceStub> {
    private AgentTypeServiceStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected AgentTypeServiceStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new AgentTypeServiceStub(channel, callOptions);
    }

    /**
     */
    public void listAgentTypes(com.park.utmstack.service.grpc.ListRequest request,
        io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.ListAgentsTypeResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getListAgentTypesMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * A stub to allow clients to do synchronous rpc calls to service AgentTypeService.
   */
  public static final class AgentTypeServiceBlockingStub
      extends io.grpc.stub.AbstractBlockingStub<AgentTypeServiceBlockingStub> {
    private AgentTypeServiceBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected AgentTypeServiceBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new AgentTypeServiceBlockingStub(channel, callOptions);
    }

    /**
     */
    public com.park.utmstack.service.grpc.ListAgentsTypeResponse listAgentTypes(com.park.utmstack.service.grpc.ListRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getListAgentTypesMethod(), getCallOptions(), request);
    }
  }

  /**
   * A stub to allow clients to do ListenableFuture-style rpc calls to service AgentTypeService.
   */
  public static final class AgentTypeServiceFutureStub
      extends io.grpc.stub.AbstractFutureStub<AgentTypeServiceFutureStub> {
    private AgentTypeServiceFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected AgentTypeServiceFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new AgentTypeServiceFutureStub(channel, callOptions);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<com.park.utmstack.service.grpc.ListAgentsTypeResponse> listAgentTypes(
        com.park.utmstack.service.grpc.ListRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getListAgentTypesMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_LIST_AGENT_TYPES = 0;

  private static final class MethodHandlers<Req, Resp> implements
      io.grpc.stub.ServerCalls.UnaryMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ServerStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ClientStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.BidiStreamingMethod<Req, Resp> {
    private final AsyncService serviceImpl;
    private final int methodId;

    MethodHandlers(AsyncService serviceImpl, int methodId) {
      this.serviceImpl = serviceImpl;
      this.methodId = methodId;
    }

    @java.lang.Override
    @java.lang.SuppressWarnings("unchecked")
    public void invoke(Req request, io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        case METHODID_LIST_AGENT_TYPES:
          serviceImpl.listAgentTypes((com.park.utmstack.service.grpc.ListRequest) request,
              (io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.ListAgentsTypeResponse>) responseObserver);
          break;
        default:
          throw new AssertionError();
      }
    }

    @java.lang.Override
    @java.lang.SuppressWarnings("unchecked")
    public io.grpc.stub.StreamObserver<Req> invoke(
        io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        default:
          throw new AssertionError();
      }
    }
  }

  public static final io.grpc.ServerServiceDefinition bindService(AsyncService service) {
    return io.grpc.ServerServiceDefinition.builder(getServiceDescriptor())
        .addMethod(
          getListAgentTypesMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              com.park.utmstack.service.grpc.ListRequest,
              com.park.utmstack.service.grpc.ListAgentsTypeResponse>(
                service, METHODID_LIST_AGENT_TYPES)))
        .build();
  }

  private static abstract class AgentTypeServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    AgentTypeServiceBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return com.park.utmstack.service.grpc.AgentManagerGrpc.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("AgentTypeService");
    }
  }

  private static final class AgentTypeServiceFileDescriptorSupplier
      extends AgentTypeServiceBaseDescriptorSupplier {
    AgentTypeServiceFileDescriptorSupplier() {}
  }

  private static final class AgentTypeServiceMethodDescriptorSupplier
      extends AgentTypeServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final String methodName;

    AgentTypeServiceMethodDescriptorSupplier(String methodName) {
      this.methodName = methodName;
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.MethodDescriptor getMethodDescriptor() {
      return getServiceDescriptor().findMethodByName(methodName);
    }
  }

  private static volatile io.grpc.ServiceDescriptor serviceDescriptor;

  public static io.grpc.ServiceDescriptor getServiceDescriptor() {
    io.grpc.ServiceDescriptor result = serviceDescriptor;
    if (result == null) {
      synchronized (AgentTypeServiceGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new AgentTypeServiceFileDescriptorSupplier())
              .addMethod(getListAgentTypesMethod())
              .build();
        }
      }
    }
    return result;
  }
}
