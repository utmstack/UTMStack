package com.park.utmstack.service.grpc;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.55.1)",
    comments = "Source: agent.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class AgentServiceGrpc {

  private AgentServiceGrpc() {}

  public static final String SERVICE_NAME = "agent.AgentService";

  // Static method descriptors that strictly reflect the proto.
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

  private static volatile io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.AgentTypeUpdate,
      com.park.utmstack.service.grpc.Agent> getUpdateAgentTypeMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "UpdateAgentType",
      requestType = com.park.utmstack.service.grpc.AgentTypeUpdate.class,
      responseType = com.park.utmstack.service.grpc.Agent.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.AgentTypeUpdate,
      com.park.utmstack.service.grpc.Agent> getUpdateAgentTypeMethod() {
    io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.AgentTypeUpdate, com.park.utmstack.service.grpc.Agent> getUpdateAgentTypeMethod;
    if ((getUpdateAgentTypeMethod = AgentServiceGrpc.getUpdateAgentTypeMethod) == null) {
      synchronized (AgentServiceGrpc.class) {
        if ((getUpdateAgentTypeMethod = AgentServiceGrpc.getUpdateAgentTypeMethod) == null) {
          AgentServiceGrpc.getUpdateAgentTypeMethod = getUpdateAgentTypeMethod =
              io.grpc.MethodDescriptor.<com.park.utmstack.service.grpc.AgentTypeUpdate, com.park.utmstack.service.grpc.Agent>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "UpdateAgentType"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.park.utmstack.service.grpc.AgentTypeUpdate.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.park.utmstack.service.grpc.Agent.getDefaultInstance()))
              .setSchemaDescriptor(new AgentServiceMethodDescriptorSupplier("UpdateAgentType"))
              .build();
        }
      }
    }
    return getUpdateAgentTypeMethod;
  }

  private static volatile io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.AgentGroupUpdate,
      com.park.utmstack.service.grpc.Agent> getUpdateAgentGroupMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "UpdateAgentGroup",
      requestType = com.park.utmstack.service.grpc.AgentGroupUpdate.class,
      responseType = com.park.utmstack.service.grpc.Agent.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.AgentGroupUpdate,
      com.park.utmstack.service.grpc.Agent> getUpdateAgentGroupMethod() {
    io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.AgentGroupUpdate, com.park.utmstack.service.grpc.Agent> getUpdateAgentGroupMethod;
    if ((getUpdateAgentGroupMethod = AgentServiceGrpc.getUpdateAgentGroupMethod) == null) {
      synchronized (AgentServiceGrpc.class) {
        if ((getUpdateAgentGroupMethod = AgentServiceGrpc.getUpdateAgentGroupMethod) == null) {
          AgentServiceGrpc.getUpdateAgentGroupMethod = getUpdateAgentGroupMethod =
              io.grpc.MethodDescriptor.<com.park.utmstack.service.grpc.AgentGroupUpdate, com.park.utmstack.service.grpc.Agent>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "UpdateAgentGroup"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.park.utmstack.service.grpc.AgentGroupUpdate.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.park.utmstack.service.grpc.Agent.getDefaultInstance()))
              .setSchemaDescriptor(new AgentServiceMethodDescriptorSupplier("UpdateAgentGroup"))
              .build();
        }
      }
    }
    return getUpdateAgentGroupMethod;
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

  private static volatile io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.Hostname,
      com.park.utmstack.service.grpc.Agent> getGetAgentByHostnameMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GetAgentByHostname",
      requestType = com.park.utmstack.service.grpc.Hostname.class,
      responseType = com.park.utmstack.service.grpc.Agent.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.Hostname,
      com.park.utmstack.service.grpc.Agent> getGetAgentByHostnameMethod() {
    io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.Hostname, com.park.utmstack.service.grpc.Agent> getGetAgentByHostnameMethod;
    if ((getGetAgentByHostnameMethod = AgentServiceGrpc.getGetAgentByHostnameMethod) == null) {
      synchronized (AgentServiceGrpc.class) {
        if ((getGetAgentByHostnameMethod = AgentServiceGrpc.getGetAgentByHostnameMethod) == null) {
          AgentServiceGrpc.getGetAgentByHostnameMethod = getGetAgentByHostnameMethod =
              io.grpc.MethodDescriptor.<com.park.utmstack.service.grpc.Hostname, com.park.utmstack.service.grpc.Agent>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GetAgentByHostname"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.park.utmstack.service.grpc.Hostname.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.park.utmstack.service.grpc.Agent.getDefaultInstance()))
              .setSchemaDescriptor(new AgentServiceMethodDescriptorSupplier("GetAgentByHostname"))
              .build();
        }
      }
    }
    return getGetAgentByHostnameMethod;
  }

  private static volatile io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.ListRequest,
      com.park.utmstack.service.grpc.ListAgentsResponse> getListAgentsWithCommandsMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ListAgentsWithCommands",
      requestType = com.park.utmstack.service.grpc.ListRequest.class,
      responseType = com.park.utmstack.service.grpc.ListAgentsResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.ListRequest,
      com.park.utmstack.service.grpc.ListAgentsResponse> getListAgentsWithCommandsMethod() {
    io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.ListRequest, com.park.utmstack.service.grpc.ListAgentsResponse> getListAgentsWithCommandsMethod;
    if ((getListAgentsWithCommandsMethod = AgentServiceGrpc.getListAgentsWithCommandsMethod) == null) {
      synchronized (AgentServiceGrpc.class) {
        if ((getListAgentsWithCommandsMethod = AgentServiceGrpc.getListAgentsWithCommandsMethod) == null) {
          AgentServiceGrpc.getListAgentsWithCommandsMethod = getListAgentsWithCommandsMethod =
              io.grpc.MethodDescriptor.<com.park.utmstack.service.grpc.ListRequest, com.park.utmstack.service.grpc.ListAgentsResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ListAgentsWithCommands"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.park.utmstack.service.grpc.ListRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.park.utmstack.service.grpc.ListAgentsResponse.getDefaultInstance()))
              .setSchemaDescriptor(new AgentServiceMethodDescriptorSupplier("ListAgentsWithCommands"))
              .build();
        }
      }
    }
    return getListAgentsWithCommandsMethod;
  }

  private static volatile io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.AgentDelete,
      com.park.utmstack.service.grpc.AgentResponse> getDeleteAgentMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "DeleteAgent",
      requestType = com.park.utmstack.service.grpc.AgentDelete.class,
      responseType = com.park.utmstack.service.grpc.AgentResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.AgentDelete,
      com.park.utmstack.service.grpc.AgentResponse> getDeleteAgentMethod() {
    io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.AgentDelete, com.park.utmstack.service.grpc.AgentResponse> getDeleteAgentMethod;
    if ((getDeleteAgentMethod = AgentServiceGrpc.getDeleteAgentMethod) == null) {
      synchronized (AgentServiceGrpc.class) {
        if ((getDeleteAgentMethod = AgentServiceGrpc.getDeleteAgentMethod) == null) {
          AgentServiceGrpc.getDeleteAgentMethod = getDeleteAgentMethod =
              io.grpc.MethodDescriptor.<com.park.utmstack.service.grpc.AgentDelete, com.park.utmstack.service.grpc.AgentResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "DeleteAgent"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.park.utmstack.service.grpc.AgentDelete.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.park.utmstack.service.grpc.AgentResponse.getDefaultInstance()))
              .setSchemaDescriptor(new AgentServiceMethodDescriptorSupplier("DeleteAgent"))
              .build();
        }
      }
    }
    return getDeleteAgentMethod;
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
    default io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.BidirectionalStream> agentStream(
        io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.BidirectionalStream> responseObserver) {
      return io.grpc.stub.ServerCalls.asyncUnimplementedStreamingCall(getAgentStreamMethod(), responseObserver);
    }

    /**
     */
    default void listAgents(com.park.utmstack.service.grpc.ListRequest request,
        io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.ListAgentsResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getListAgentsMethod(), responseObserver);
    }

    /**
     */
    default void updateAgentType(com.park.utmstack.service.grpc.AgentTypeUpdate request,
        io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.Agent> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getUpdateAgentTypeMethod(), responseObserver);
    }

    /**
     */
    default void updateAgentGroup(com.park.utmstack.service.grpc.AgentGroupUpdate request,
        io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.Agent> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getUpdateAgentGroupMethod(), responseObserver);
    }

    /**
     */
    default void listAgentCommands(com.park.utmstack.service.grpc.ListRequest request,
        io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.ListAgentsCommandsResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getListAgentCommandsMethod(), responseObserver);
    }

    /**
     */
    default void getAgentByHostname(com.park.utmstack.service.grpc.Hostname request,
        io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.Agent> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGetAgentByHostnameMethod(), responseObserver);
    }

    /**
     */
    default void listAgentsWithCommands(com.park.utmstack.service.grpc.ListRequest request,
        io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.ListAgentsResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getListAgentsWithCommandsMethod(), responseObserver);
    }

    /**
     */
    default void deleteAgent(com.park.utmstack.service.grpc.AgentDelete request,
        io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.AgentResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getDeleteAgentMethod(), responseObserver);
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
    public io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.BidirectionalStream> agentStream(
        io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.BidirectionalStream> responseObserver) {
      return io.grpc.stub.ClientCalls.asyncBidiStreamingCall(
          getChannel().newCall(getAgentStreamMethod(), getCallOptions()), responseObserver);
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
    public void updateAgentType(com.park.utmstack.service.grpc.AgentTypeUpdate request,
        io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.Agent> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getUpdateAgentTypeMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void updateAgentGroup(com.park.utmstack.service.grpc.AgentGroupUpdate request,
        io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.Agent> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getUpdateAgentGroupMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void listAgentCommands(com.park.utmstack.service.grpc.ListRequest request,
        io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.ListAgentsCommandsResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getListAgentCommandsMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void getAgentByHostname(com.park.utmstack.service.grpc.Hostname request,
        io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.Agent> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGetAgentByHostnameMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void listAgentsWithCommands(com.park.utmstack.service.grpc.ListRequest request,
        io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.ListAgentsResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getListAgentsWithCommandsMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void deleteAgent(com.park.utmstack.service.grpc.AgentDelete request,
        io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.AgentResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getDeleteAgentMethod(), getCallOptions()), request, responseObserver);
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
    public com.park.utmstack.service.grpc.ListAgentsResponse listAgents(com.park.utmstack.service.grpc.ListRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getListAgentsMethod(), getCallOptions(), request);
    }

    /**
     */
    public com.park.utmstack.service.grpc.Agent updateAgentType(com.park.utmstack.service.grpc.AgentTypeUpdate request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getUpdateAgentTypeMethod(), getCallOptions(), request);
    }

    /**
     */
    public com.park.utmstack.service.grpc.Agent updateAgentGroup(com.park.utmstack.service.grpc.AgentGroupUpdate request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getUpdateAgentGroupMethod(), getCallOptions(), request);
    }

    /**
     */
    public com.park.utmstack.service.grpc.ListAgentsCommandsResponse listAgentCommands(com.park.utmstack.service.grpc.ListRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getListAgentCommandsMethod(), getCallOptions(), request);
    }

    /**
     */
    public com.park.utmstack.service.grpc.Agent getAgentByHostname(com.park.utmstack.service.grpc.Hostname request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGetAgentByHostnameMethod(), getCallOptions(), request);
    }

    /**
     */
    public com.park.utmstack.service.grpc.ListAgentsResponse listAgentsWithCommands(com.park.utmstack.service.grpc.ListRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getListAgentsWithCommandsMethod(), getCallOptions(), request);
    }

    /**
     */
    public com.park.utmstack.service.grpc.AgentResponse deleteAgent(com.park.utmstack.service.grpc.AgentDelete request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getDeleteAgentMethod(), getCallOptions(), request);
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
    public com.google.common.util.concurrent.ListenableFuture<com.park.utmstack.service.grpc.ListAgentsResponse> listAgents(
        com.park.utmstack.service.grpc.ListRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getListAgentsMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<com.park.utmstack.service.grpc.Agent> updateAgentType(
        com.park.utmstack.service.grpc.AgentTypeUpdate request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getUpdateAgentTypeMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<com.park.utmstack.service.grpc.Agent> updateAgentGroup(
        com.park.utmstack.service.grpc.AgentGroupUpdate request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getUpdateAgentGroupMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<com.park.utmstack.service.grpc.ListAgentsCommandsResponse> listAgentCommands(
        com.park.utmstack.service.grpc.ListRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getListAgentCommandsMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<com.park.utmstack.service.grpc.Agent> getAgentByHostname(
        com.park.utmstack.service.grpc.Hostname request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGetAgentByHostnameMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<com.park.utmstack.service.grpc.ListAgentsResponse> listAgentsWithCommands(
        com.park.utmstack.service.grpc.ListRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getListAgentsWithCommandsMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<com.park.utmstack.service.grpc.AgentResponse> deleteAgent(
        com.park.utmstack.service.grpc.AgentDelete request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getDeleteAgentMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_LIST_AGENTS = 0;
  private static final int METHODID_UPDATE_AGENT_TYPE = 1;
  private static final int METHODID_UPDATE_AGENT_GROUP = 2;
  private static final int METHODID_LIST_AGENT_COMMANDS = 3;
  private static final int METHODID_GET_AGENT_BY_HOSTNAME = 4;
  private static final int METHODID_LIST_AGENTS_WITH_COMMANDS = 5;
  private static final int METHODID_DELETE_AGENT = 6;
  private static final int METHODID_AGENT_STREAM = 7;

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
        case METHODID_LIST_AGENTS:
          serviceImpl.listAgents((com.park.utmstack.service.grpc.ListRequest) request,
              (io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.ListAgentsResponse>) responseObserver);
          break;
        case METHODID_UPDATE_AGENT_TYPE:
          serviceImpl.updateAgentType((com.park.utmstack.service.grpc.AgentTypeUpdate) request,
              (io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.Agent>) responseObserver);
          break;
        case METHODID_UPDATE_AGENT_GROUP:
          serviceImpl.updateAgentGroup((com.park.utmstack.service.grpc.AgentGroupUpdate) request,
              (io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.Agent>) responseObserver);
          break;
        case METHODID_LIST_AGENT_COMMANDS:
          serviceImpl.listAgentCommands((com.park.utmstack.service.grpc.ListRequest) request,
              (io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.ListAgentsCommandsResponse>) responseObserver);
          break;
        case METHODID_GET_AGENT_BY_HOSTNAME:
          serviceImpl.getAgentByHostname((com.park.utmstack.service.grpc.Hostname) request,
              (io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.Agent>) responseObserver);
          break;
        case METHODID_LIST_AGENTS_WITH_COMMANDS:
          serviceImpl.listAgentsWithCommands((com.park.utmstack.service.grpc.ListRequest) request,
              (io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.ListAgentsResponse>) responseObserver);
          break;
        case METHODID_DELETE_AGENT:
          serviceImpl.deleteAgent((com.park.utmstack.service.grpc.AgentDelete) request,
              (io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.AgentResponse>) responseObserver);
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
          getAgentStreamMethod(),
          io.grpc.stub.ServerCalls.asyncBidiStreamingCall(
            new MethodHandlers<
              com.park.utmstack.service.grpc.BidirectionalStream,
              com.park.utmstack.service.grpc.BidirectionalStream>(
                service, METHODID_AGENT_STREAM)))
        .addMethod(
          getListAgentsMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              com.park.utmstack.service.grpc.ListRequest,
              com.park.utmstack.service.grpc.ListAgentsResponse>(
                service, METHODID_LIST_AGENTS)))
        .addMethod(
          getUpdateAgentTypeMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              com.park.utmstack.service.grpc.AgentTypeUpdate,
              com.park.utmstack.service.grpc.Agent>(
                service, METHODID_UPDATE_AGENT_TYPE)))
        .addMethod(
          getUpdateAgentGroupMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              com.park.utmstack.service.grpc.AgentGroupUpdate,
              com.park.utmstack.service.grpc.Agent>(
                service, METHODID_UPDATE_AGENT_GROUP)))
        .addMethod(
          getListAgentCommandsMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              com.park.utmstack.service.grpc.ListRequest,
              com.park.utmstack.service.grpc.ListAgentsCommandsResponse>(
                service, METHODID_LIST_AGENT_COMMANDS)))
        .addMethod(
          getGetAgentByHostnameMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              com.park.utmstack.service.grpc.Hostname,
              com.park.utmstack.service.grpc.Agent>(
                service, METHODID_GET_AGENT_BY_HOSTNAME)))
        .addMethod(
          getListAgentsWithCommandsMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              com.park.utmstack.service.grpc.ListRequest,
              com.park.utmstack.service.grpc.ListAgentsResponse>(
                service, METHODID_LIST_AGENTS_WITH_COMMANDS)))
        .addMethod(
          getDeleteAgentMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              com.park.utmstack.service.grpc.AgentDelete,
              com.park.utmstack.service.grpc.AgentResponse>(
                service, METHODID_DELETE_AGENT)))
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
    private final String methodName;

    AgentServiceMethodDescriptorSupplier(String methodName) {
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
              .addMethod(getAgentStreamMethod())
              .addMethod(getListAgentsMethod())
              .addMethod(getUpdateAgentTypeMethod())
              .addMethod(getUpdateAgentGroupMethod())
              .addMethod(getListAgentCommandsMethod())
              .addMethod(getGetAgentByHostnameMethod())
              .addMethod(getListAgentsWithCommandsMethod())
              .addMethod(getDeleteAgentMethod())
              .build();
        }
      }
    }
    return result;
  }
}
