package com.park.utmstack.service.grpc;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.65.1)",
    comments = "Source: agent.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class AgentServiceGrpc {

  private AgentServiceGrpc() {}

  public static final java.lang.String SERVICE_NAME = "agent.AgentService";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.AgentRequest,
      com.park.utmstack.service.grpc.AuthResponse> getRegisterAgentMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "RegisterAgent",
      requestType = com.park.utmstack.service.grpc.AgentRequest.class,
      responseType = com.park.utmstack.service.grpc.AuthResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.AgentRequest,
      com.park.utmstack.service.grpc.AuthResponse> getRegisterAgentMethod() {
    io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.AgentRequest, com.park.utmstack.service.grpc.AuthResponse> getRegisterAgentMethod;
    if ((getRegisterAgentMethod = AgentServiceGrpc.getRegisterAgentMethod) == null) {
      synchronized (AgentServiceGrpc.class) {
        if ((getRegisterAgentMethod = AgentServiceGrpc.getRegisterAgentMethod) == null) {
          AgentServiceGrpc.getRegisterAgentMethod = getRegisterAgentMethod =
              io.grpc.MethodDescriptor.<com.park.utmstack.service.grpc.AgentRequest, com.park.utmstack.service.grpc.AuthResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "RegisterAgent"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.park.utmstack.service.grpc.AgentRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.park.utmstack.service.grpc.AuthResponse.getDefaultInstance()))
              .setSchemaDescriptor(new AgentServiceMethodDescriptorSupplier("RegisterAgent"))
              .build();
        }
      }
    }
    return getRegisterAgentMethod;
  }

  private static volatile io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.DeleteRequest,
      com.park.utmstack.service.grpc.AuthResponse> getDeleteAgentMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "DeleteAgent",
      requestType = com.park.utmstack.service.grpc.DeleteRequest.class,
      responseType = com.park.utmstack.service.grpc.AuthResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.DeleteRequest,
      com.park.utmstack.service.grpc.AuthResponse> getDeleteAgentMethod() {
    io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.DeleteRequest, com.park.utmstack.service.grpc.AuthResponse> getDeleteAgentMethod;
    if ((getDeleteAgentMethod = AgentServiceGrpc.getDeleteAgentMethod) == null) {
      synchronized (AgentServiceGrpc.class) {
        if ((getDeleteAgentMethod = AgentServiceGrpc.getDeleteAgentMethod) == null) {
          AgentServiceGrpc.getDeleteAgentMethod = getDeleteAgentMethod =
              io.grpc.MethodDescriptor.<com.park.utmstack.service.grpc.DeleteRequest, com.park.utmstack.service.grpc.AuthResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "DeleteAgent"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.park.utmstack.service.grpc.DeleteRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.park.utmstack.service.grpc.AuthResponse.getDefaultInstance()))
              .setSchemaDescriptor(new AgentServiceMethodDescriptorSupplier("DeleteAgent"))
              .build();
        }
      }
    }
    return getDeleteAgentMethod;
  }

  private static volatile io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.ListRequest,
      com.park.utmstack.service.grpc.ListAgentsResponse> getListAgentsMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ListAgents",
      requestType = com.park.utmstack.service.grpc.ListRequest.class,
      responseType = com.park.utmstack.service.grpc.ListAgentsResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.ListRequest,
      com.park.utmstack.service.grpc.ListAgentsResponse> getListAgentsMethod() {
    io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.ListRequest, com.park.utmstack.service.grpc.ListAgentsResponse> getListAgentsMethod;
    if ((getListAgentsMethod = AgentServiceGrpc.getListAgentsMethod) == null) {
      synchronized (AgentServiceGrpc.class) {
        if ((getListAgentsMethod = AgentServiceGrpc.getListAgentsMethod) == null) {
          AgentServiceGrpc.getListAgentsMethod = getListAgentsMethod =
              io.grpc.MethodDescriptor.<com.park.utmstack.service.grpc.ListRequest, com.park.utmstack.service.grpc.ListAgentsResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ListAgents"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.park.utmstack.service.grpc.ListRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.park.utmstack.service.grpc.ListAgentsResponse.getDefaultInstance()))
              .setSchemaDescriptor(new AgentServiceMethodDescriptorSupplier("ListAgents"))
              .build();
        }
      }
    }
    return getListAgentsMethod;
  }

  private static volatile io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.BidirectionalStream,
      com.park.utmstack.service.grpc.BidirectionalStream> getAgentStreamMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "AgentStream",
      requestType = com.park.utmstack.service.grpc.BidirectionalStream.class,
      responseType = com.park.utmstack.service.grpc.BidirectionalStream.class,
      methodType = io.grpc.MethodDescriptor.MethodType.BIDI_STREAMING)
  public static io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.BidirectionalStream,
      com.park.utmstack.service.grpc.BidirectionalStream> getAgentStreamMethod() {
    io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.BidirectionalStream, com.park.utmstack.service.grpc.BidirectionalStream> getAgentStreamMethod;
    if ((getAgentStreamMethod = AgentServiceGrpc.getAgentStreamMethod) == null) {
      synchronized (AgentServiceGrpc.class) {
        if ((getAgentStreamMethod = AgentServiceGrpc.getAgentStreamMethod) == null) {
          AgentServiceGrpc.getAgentStreamMethod = getAgentStreamMethod =
              io.grpc.MethodDescriptor.<com.park.utmstack.service.grpc.BidirectionalStream, com.park.utmstack.service.grpc.BidirectionalStream>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.BIDI_STREAMING)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "AgentStream"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.park.utmstack.service.grpc.BidirectionalStream.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.park.utmstack.service.grpc.BidirectionalStream.getDefaultInstance()))
              .setSchemaDescriptor(new AgentServiceMethodDescriptorSupplier("AgentStream"))
              .build();
        }
      }
    }
    return getAgentStreamMethod;
  }

  private static volatile io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.ListRequest,
      com.park.utmstack.service.grpc.ListAgentsCommandsResponse> getListAgentCommandsMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ListAgentCommands",
      requestType = com.park.utmstack.service.grpc.ListRequest.class,
      responseType = com.park.utmstack.service.grpc.ListAgentsCommandsResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.ListRequest,
      com.park.utmstack.service.grpc.ListAgentsCommandsResponse> getListAgentCommandsMethod() {
    io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.ListRequest, com.park.utmstack.service.grpc.ListAgentsCommandsResponse> getListAgentCommandsMethod;
    if ((getListAgentCommandsMethod = AgentServiceGrpc.getListAgentCommandsMethod) == null) {
      synchronized (AgentServiceGrpc.class) {
        if ((getListAgentCommandsMethod = AgentServiceGrpc.getListAgentCommandsMethod) == null) {
          AgentServiceGrpc.getListAgentCommandsMethod = getListAgentCommandsMethod =
              io.grpc.MethodDescriptor.<com.park.utmstack.service.grpc.ListRequest, com.park.utmstack.service.grpc.ListAgentsCommandsResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ListAgentCommands"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.park.utmstack.service.grpc.ListRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.park.utmstack.service.grpc.ListAgentsCommandsResponse.getDefaultInstance()))
              .setSchemaDescriptor(new AgentServiceMethodDescriptorSupplier("ListAgentCommands"))
              .build();
        }
      }
    }
    return getListAgentCommandsMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static AgentServiceStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<AgentServiceStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<AgentServiceStub>() {
        @java.lang.Override
        public AgentServiceStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new AgentServiceStub(channel, callOptions);
        }
      };
    return AgentServiceStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static AgentServiceBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<AgentServiceBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<AgentServiceBlockingStub>() {
        @java.lang.Override
        public AgentServiceBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new AgentServiceBlockingStub(channel, callOptions);
        }
      };
    return AgentServiceBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static AgentServiceFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<AgentServiceFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<AgentServiceFutureStub>() {
        @java.lang.Override
        public AgentServiceFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new AgentServiceFutureStub(channel, callOptions);
        }
      };
    return AgentServiceFutureStub.newStub(factory, channel);
  }

  /**
   */
  public interface AsyncService {

    /**
     */
    default void registerAgent(com.park.utmstack.service.grpc.AgentRequest request,
        io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.AuthResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getRegisterAgentMethod(), responseObserver);
    }

    /**
     */
    default void deleteAgent(com.park.utmstack.service.grpc.DeleteRequest request,
        io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.AuthResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getDeleteAgentMethod(), responseObserver);
    }

    /**
     */
    default void listAgents(com.park.utmstack.service.grpc.ListRequest request,
        io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.ListAgentsResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getListAgentsMethod(), responseObserver);
    }

    /**
     */
    default io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.BidirectionalStream> agentStream(
        io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.BidirectionalStream> responseObserver) {
      return io.grpc.stub.ServerCalls.asyncUnimplementedStreamingCall(getAgentStreamMethod(), responseObserver);
    }

    /**
     */
    default void listAgentCommands(com.park.utmstack.service.grpc.ListRequest request,
        io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.ListAgentsCommandsResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getListAgentCommandsMethod(), responseObserver);
    }
  }

  /**
   * Base class for the server implementation of the service AgentService.
   */
  public static abstract class AgentServiceImplBase
      implements io.grpc.BindableService, AsyncService {

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return AgentServiceGrpc.bindService(this);
    }
  }

  /**
   * A stub to allow clients to do asynchronous rpc calls to service AgentService.
   */
  public static final class AgentServiceStub
      extends io.grpc.stub.AbstractAsyncStub<AgentServiceStub> {
    private AgentServiceStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected AgentServiceStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new AgentServiceStub(channel, callOptions);
    }

    /**
     */
    public void registerAgent(com.park.utmstack.service.grpc.AgentRequest request,
        io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.AuthResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getRegisterAgentMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void deleteAgent(com.park.utmstack.service.grpc.DeleteRequest request,
        io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.AuthResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getDeleteAgentMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void listAgents(com.park.utmstack.service.grpc.ListRequest request,
        io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.ListAgentsResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getListAgentsMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.BidirectionalStream> agentStream(
        io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.BidirectionalStream> responseObserver) {
      return io.grpc.stub.ClientCalls.asyncBidiStreamingCall(
          getChannel().newCall(getAgentStreamMethod(), getCallOptions()), responseObserver);
    }

    /**
     */
    public void listAgentCommands(com.park.utmstack.service.grpc.ListRequest request,
        io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.ListAgentsCommandsResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getListAgentCommandsMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * A stub to allow clients to do synchronous rpc calls to service AgentService.
   */
  public static final class AgentServiceBlockingStub
      extends io.grpc.stub.AbstractBlockingStub<AgentServiceBlockingStub> {
    private AgentServiceBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected AgentServiceBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new AgentServiceBlockingStub(channel, callOptions);
    }

    /**
     */
    public com.park.utmstack.service.grpc.AuthResponse registerAgent(com.park.utmstack.service.grpc.AgentRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getRegisterAgentMethod(), getCallOptions(), request);
    }

    /**
     */
    public com.park.utmstack.service.grpc.AuthResponse deleteAgent(com.park.utmstack.service.grpc.DeleteRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getDeleteAgentMethod(), getCallOptions(), request);
    }

    /**
     */
    public com.park.utmstack.service.grpc.ListAgentsResponse listAgents(com.park.utmstack.service.grpc.ListRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getListAgentsMethod(), getCallOptions(), request);
    }

    /**
     */
    public com.park.utmstack.service.grpc.ListAgentsCommandsResponse listAgentCommands(com.park.utmstack.service.grpc.ListRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getListAgentCommandsMethod(), getCallOptions(), request);
    }
  }

  /**
   * A stub to allow clients to do ListenableFuture-style rpc calls to service AgentService.
   */
  public static final class AgentServiceFutureStub
      extends io.grpc.stub.AbstractFutureStub<AgentServiceFutureStub> {
    private AgentServiceFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected AgentServiceFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new AgentServiceFutureStub(channel, callOptions);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<com.park.utmstack.service.grpc.AuthResponse> registerAgent(
        com.park.utmstack.service.grpc.AgentRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getRegisterAgentMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<com.park.utmstack.service.grpc.AuthResponse> deleteAgent(
        com.park.utmstack.service.grpc.DeleteRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getDeleteAgentMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<com.park.utmstack.service.grpc.ListAgentsResponse> listAgents(
        com.park.utmstack.service.grpc.ListRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getListAgentsMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<com.park.utmstack.service.grpc.ListAgentsCommandsResponse> listAgentCommands(
        com.park.utmstack.service.grpc.ListRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getListAgentCommandsMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_REGISTER_AGENT = 0;
  private static final int METHODID_DELETE_AGENT = 1;
  private static final int METHODID_LIST_AGENTS = 2;
  private static final int METHODID_LIST_AGENT_COMMANDS = 3;
  private static final int METHODID_AGENT_STREAM = 4;

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
        case METHODID_REGISTER_AGENT:
          serviceImpl.registerAgent((com.park.utmstack.service.grpc.AgentRequest) request,
              (io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.AuthResponse>) responseObserver);
          break;
        case METHODID_DELETE_AGENT:
          serviceImpl.deleteAgent((com.park.utmstack.service.grpc.DeleteRequest) request,
              (io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.AuthResponse>) responseObserver);
          break;
        case METHODID_LIST_AGENTS:
          serviceImpl.listAgents((com.park.utmstack.service.grpc.ListRequest) request,
              (io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.ListAgentsResponse>) responseObserver);
          break;
        case METHODID_LIST_AGENT_COMMANDS:
          serviceImpl.listAgentCommands((com.park.utmstack.service.grpc.ListRequest) request,
              (io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.ListAgentsCommandsResponse>) responseObserver);
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
        case METHODID_AGENT_STREAM:
          return (io.grpc.stub.StreamObserver<Req>) serviceImpl.agentStream(
              (io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.BidirectionalStream>) responseObserver);
        default:
          throw new AssertionError();
      }
    }
  }

  public static final io.grpc.ServerServiceDefinition bindService(AsyncService service) {
    return io.grpc.ServerServiceDefinition.builder(getServiceDescriptor())
        .addMethod(
          getRegisterAgentMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              com.park.utmstack.service.grpc.AgentRequest,
              com.park.utmstack.service.grpc.AuthResponse>(
                service, METHODID_REGISTER_AGENT)))
        .addMethod(
          getDeleteAgentMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              com.park.utmstack.service.grpc.DeleteRequest,
              com.park.utmstack.service.grpc.AuthResponse>(
                service, METHODID_DELETE_AGENT)))
        .addMethod(
          getListAgentsMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              com.park.utmstack.service.grpc.ListRequest,
              com.park.utmstack.service.grpc.ListAgentsResponse>(
                service, METHODID_LIST_AGENTS)))
        .addMethod(
          getAgentStreamMethod(),
          io.grpc.stub.ServerCalls.asyncBidiStreamingCall(
            new MethodHandlers<
              com.park.utmstack.service.grpc.BidirectionalStream,
              com.park.utmstack.service.grpc.BidirectionalStream>(
                service, METHODID_AGENT_STREAM)))
        .addMethod(
          getListAgentCommandsMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              com.park.utmstack.service.grpc.ListRequest,
              com.park.utmstack.service.grpc.ListAgentsCommandsResponse>(
                service, METHODID_LIST_AGENT_COMMANDS)))
        .build();
  }

  private static abstract class AgentServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    AgentServiceBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return com.park.utmstack.service.grpc.AgentManagerGrpc.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("AgentService");
    }
  }

  private static final class AgentServiceFileDescriptorSupplier
      extends AgentServiceBaseDescriptorSupplier {
    AgentServiceFileDescriptorSupplier() {}
  }

  private static final class AgentServiceMethodDescriptorSupplier
      extends AgentServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final java.lang.String methodName;

    AgentServiceMethodDescriptorSupplier(java.lang.String methodName) {
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
      synchronized (AgentServiceGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new AgentServiceFileDescriptorSupplier())
              .addMethod(getRegisterAgentMethod())
              .addMethod(getDeleteAgentMethod())
              .addMethod(getListAgentsMethod())
              .addMethod(getAgentStreamMethod())
              .addMethod(getListAgentCommandsMethod())
              .build();
        }
      }
    }
    return result;
  }
}
