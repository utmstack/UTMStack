package com.park.utmstack.service.grpc;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.55.1)",
    comments = "Source: agent.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class AgentGroupServiceGrpc {

  private AgentGroupServiceGrpc() {}

  public static final String SERVICE_NAME = "agent.AgentGroupService";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.AgentGroup,
      com.park.utmstack.service.grpc.AgentGroup> getCreateGroupMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "CreateGroup",
      requestType = com.park.utmstack.service.grpc.AgentGroup.class,
      responseType = com.park.utmstack.service.grpc.AgentGroup.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.AgentGroup,
      com.park.utmstack.service.grpc.AgentGroup> getCreateGroupMethod() {
    io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.AgentGroup, com.park.utmstack.service.grpc.AgentGroup> getCreateGroupMethod;
    if ((getCreateGroupMethod = AgentGroupServiceGrpc.getCreateGroupMethod) == null) {
      synchronized (AgentGroupServiceGrpc.class) {
        if ((getCreateGroupMethod = AgentGroupServiceGrpc.getCreateGroupMethod) == null) {
          AgentGroupServiceGrpc.getCreateGroupMethod = getCreateGroupMethod =
              io.grpc.MethodDescriptor.<com.park.utmstack.service.grpc.AgentGroup, com.park.utmstack.service.grpc.AgentGroup>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "CreateGroup"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.park.utmstack.service.grpc.AgentGroup.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.park.utmstack.service.grpc.AgentGroup.getDefaultInstance()))
              .setSchemaDescriptor(new AgentGroupServiceMethodDescriptorSupplier("CreateGroup"))
              .build();
        }
      }
    }
    return getCreateGroupMethod;
  }

  private static volatile io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.AgentGroup,
      com.park.utmstack.service.grpc.AgentGroup> getEditGroupMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "EditGroup",
      requestType = com.park.utmstack.service.grpc.AgentGroup.class,
      responseType = com.park.utmstack.service.grpc.AgentGroup.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.AgentGroup,
      com.park.utmstack.service.grpc.AgentGroup> getEditGroupMethod() {
    io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.AgentGroup, com.park.utmstack.service.grpc.AgentGroup> getEditGroupMethod;
    if ((getEditGroupMethod = AgentGroupServiceGrpc.getEditGroupMethod) == null) {
      synchronized (AgentGroupServiceGrpc.class) {
        if ((getEditGroupMethod = AgentGroupServiceGrpc.getEditGroupMethod) == null) {
          AgentGroupServiceGrpc.getEditGroupMethod = getEditGroupMethod =
              io.grpc.MethodDescriptor.<com.park.utmstack.service.grpc.AgentGroup, com.park.utmstack.service.grpc.AgentGroup>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "EditGroup"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.park.utmstack.service.grpc.AgentGroup.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.park.utmstack.service.grpc.AgentGroup.getDefaultInstance()))
              .setSchemaDescriptor(new AgentGroupServiceMethodDescriptorSupplier("EditGroup"))
              .build();
        }
      }
    }
    return getEditGroupMethod;
  }

  private static volatile io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.ListRequest,
      com.park.utmstack.service.grpc.ListAgentsGroupResponse> getListGroupsMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ListGroups",
      requestType = com.park.utmstack.service.grpc.ListRequest.class,
      responseType = com.park.utmstack.service.grpc.ListAgentsGroupResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.ListRequest,
      com.park.utmstack.service.grpc.ListAgentsGroupResponse> getListGroupsMethod() {
    io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.ListRequest, com.park.utmstack.service.grpc.ListAgentsGroupResponse> getListGroupsMethod;
    if ((getListGroupsMethod = AgentGroupServiceGrpc.getListGroupsMethod) == null) {
      synchronized (AgentGroupServiceGrpc.class) {
        if ((getListGroupsMethod = AgentGroupServiceGrpc.getListGroupsMethod) == null) {
          AgentGroupServiceGrpc.getListGroupsMethod = getListGroupsMethod =
              io.grpc.MethodDescriptor.<com.park.utmstack.service.grpc.ListRequest, com.park.utmstack.service.grpc.ListAgentsGroupResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ListGroups"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.park.utmstack.service.grpc.ListRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.park.utmstack.service.grpc.ListAgentsGroupResponse.getDefaultInstance()))
              .setSchemaDescriptor(new AgentGroupServiceMethodDescriptorSupplier("ListGroups"))
              .build();
        }
      }
    }
    return getListGroupsMethod;
  }

  private static volatile io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.Id,
      com.park.utmstack.service.grpc.Id> getDeleteGroupMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "DeleteGroup",
      requestType = com.park.utmstack.service.grpc.Id.class,
      responseType = com.park.utmstack.service.grpc.Id.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.Id,
      com.park.utmstack.service.grpc.Id> getDeleteGroupMethod() {
    io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.Id, com.park.utmstack.service.grpc.Id> getDeleteGroupMethod;
    if ((getDeleteGroupMethod = AgentGroupServiceGrpc.getDeleteGroupMethod) == null) {
      synchronized (AgentGroupServiceGrpc.class) {
        if ((getDeleteGroupMethod = AgentGroupServiceGrpc.getDeleteGroupMethod) == null) {
          AgentGroupServiceGrpc.getDeleteGroupMethod = getDeleteGroupMethod =
              io.grpc.MethodDescriptor.<com.park.utmstack.service.grpc.Id, com.park.utmstack.service.grpc.Id>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "DeleteGroup"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.park.utmstack.service.grpc.Id.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.park.utmstack.service.grpc.Id.getDefaultInstance()))
              .setSchemaDescriptor(new AgentGroupServiceMethodDescriptorSupplier("DeleteGroup"))
              .build();
        }
      }
    }
    return getDeleteGroupMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static AgentGroupServiceStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<AgentGroupServiceStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<AgentGroupServiceStub>() {
        @java.lang.Override
        public AgentGroupServiceStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new AgentGroupServiceStub(channel, callOptions);
        }
      };
    return AgentGroupServiceStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static AgentGroupServiceBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<AgentGroupServiceBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<AgentGroupServiceBlockingStub>() {
        @java.lang.Override
        public AgentGroupServiceBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new AgentGroupServiceBlockingStub(channel, callOptions);
        }
      };
    return AgentGroupServiceBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static AgentGroupServiceFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<AgentGroupServiceFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<AgentGroupServiceFutureStub>() {
        @java.lang.Override
        public AgentGroupServiceFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new AgentGroupServiceFutureStub(channel, callOptions);
        }
      };
    return AgentGroupServiceFutureStub.newStub(factory, channel);
  }

  /**
   */
  public interface AsyncService {

    /**
     */
    default void createGroup(com.park.utmstack.service.grpc.AgentGroup request,
        io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.AgentGroup> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getCreateGroupMethod(), responseObserver);
    }

    /**
     */
    default void editGroup(com.park.utmstack.service.grpc.AgentGroup request,
        io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.AgentGroup> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getEditGroupMethod(), responseObserver);
    }

    /**
     */
    default void listGroups(com.park.utmstack.service.grpc.ListRequest request,
        io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.ListAgentsGroupResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getListGroupsMethod(), responseObserver);
    }

    /**
     */
    default void deleteGroup(com.park.utmstack.service.grpc.Id request,
        io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.Id> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getDeleteGroupMethod(), responseObserver);
    }
  }

  /**
   * Base class for the server implementation of the service AgentGroupService.
   */
  public static abstract class AgentGroupServiceImplBase
      implements io.grpc.BindableService, AsyncService {

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return AgentGroupServiceGrpc.bindService(this);
    }
  }

  /**
   * A stub to allow clients to do asynchronous rpc calls to service AgentGroupService.
   */
  public static final class AgentGroupServiceStub
      extends io.grpc.stub.AbstractAsyncStub<AgentGroupServiceStub> {
    private AgentGroupServiceStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected AgentGroupServiceStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new AgentGroupServiceStub(channel, callOptions);
    }

    /**
     */
    public void createGroup(com.park.utmstack.service.grpc.AgentGroup request,
        io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.AgentGroup> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getCreateGroupMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void editGroup(com.park.utmstack.service.grpc.AgentGroup request,
        io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.AgentGroup> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getEditGroupMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void listGroups(com.park.utmstack.service.grpc.ListRequest request,
        io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.ListAgentsGroupResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getListGroupsMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void deleteGroup(com.park.utmstack.service.grpc.Id request,
        io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.Id> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getDeleteGroupMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * A stub to allow clients to do synchronous rpc calls to service AgentGroupService.
   */
  public static final class AgentGroupServiceBlockingStub
      extends io.grpc.stub.AbstractBlockingStub<AgentGroupServiceBlockingStub> {
    private AgentGroupServiceBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected AgentGroupServiceBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new AgentGroupServiceBlockingStub(channel, callOptions);
    }

    /**
     */
    public com.park.utmstack.service.grpc.AgentGroup createGroup(com.park.utmstack.service.grpc.AgentGroup request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getCreateGroupMethod(), getCallOptions(), request);
    }

    /**
     */
    public com.park.utmstack.service.grpc.AgentGroup editGroup(com.park.utmstack.service.grpc.AgentGroup request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getEditGroupMethod(), getCallOptions(), request);
    }

    /**
     */
    public com.park.utmstack.service.grpc.ListAgentsGroupResponse listGroups(com.park.utmstack.service.grpc.ListRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getListGroupsMethod(), getCallOptions(), request);
    }

    /**
     */
    public com.park.utmstack.service.grpc.Id deleteGroup(com.park.utmstack.service.grpc.Id request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getDeleteGroupMethod(), getCallOptions(), request);
    }
  }

  /**
   * A stub to allow clients to do ListenableFuture-style rpc calls to service AgentGroupService.
   */
  public static final class AgentGroupServiceFutureStub
      extends io.grpc.stub.AbstractFutureStub<AgentGroupServiceFutureStub> {
    private AgentGroupServiceFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected AgentGroupServiceFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new AgentGroupServiceFutureStub(channel, callOptions);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<com.park.utmstack.service.grpc.AgentGroup> createGroup(
        com.park.utmstack.service.grpc.AgentGroup request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getCreateGroupMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<com.park.utmstack.service.grpc.AgentGroup> editGroup(
        com.park.utmstack.service.grpc.AgentGroup request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getEditGroupMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<com.park.utmstack.service.grpc.ListAgentsGroupResponse> listGroups(
        com.park.utmstack.service.grpc.ListRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getListGroupsMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<com.park.utmstack.service.grpc.Id> deleteGroup(
        com.park.utmstack.service.grpc.Id request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getDeleteGroupMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_CREATE_GROUP = 0;
  private static final int METHODID_EDIT_GROUP = 1;
  private static final int METHODID_LIST_GROUPS = 2;
  private static final int METHODID_DELETE_GROUP = 3;

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
        case METHODID_CREATE_GROUP:
          serviceImpl.createGroup((com.park.utmstack.service.grpc.AgentGroup) request,
              (io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.AgentGroup>) responseObserver);
          break;
        case METHODID_EDIT_GROUP:
          serviceImpl.editGroup((com.park.utmstack.service.grpc.AgentGroup) request,
              (io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.AgentGroup>) responseObserver);
          break;
        case METHODID_LIST_GROUPS:
          serviceImpl.listGroups((com.park.utmstack.service.grpc.ListRequest) request,
              (io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.ListAgentsGroupResponse>) responseObserver);
          break;
        case METHODID_DELETE_GROUP:
          serviceImpl.deleteGroup((com.park.utmstack.service.grpc.Id) request,
              (io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.Id>) responseObserver);
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
          getCreateGroupMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              com.park.utmstack.service.grpc.AgentGroup,
              com.park.utmstack.service.grpc.AgentGroup>(
                service, METHODID_CREATE_GROUP)))
        .addMethod(
          getEditGroupMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              com.park.utmstack.service.grpc.AgentGroup,
              com.park.utmstack.service.grpc.AgentGroup>(
                service, METHODID_EDIT_GROUP)))
        .addMethod(
          getListGroupsMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              com.park.utmstack.service.grpc.ListRequest,
              com.park.utmstack.service.grpc.ListAgentsGroupResponse>(
                service, METHODID_LIST_GROUPS)))
        .addMethod(
          getDeleteGroupMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              com.park.utmstack.service.grpc.Id,
              com.park.utmstack.service.grpc.Id>(
                service, METHODID_DELETE_GROUP)))
        .build();
  }

  private static abstract class AgentGroupServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    AgentGroupServiceBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return com.park.utmstack.service.grpc.AgentManagerGrpc.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("AgentGroupService");
    }
  }

  private static final class AgentGroupServiceFileDescriptorSupplier
      extends AgentGroupServiceBaseDescriptorSupplier {
    AgentGroupServiceFileDescriptorSupplier() {}
  }

  private static final class AgentGroupServiceMethodDescriptorSupplier
      extends AgentGroupServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final String methodName;

    AgentGroupServiceMethodDescriptorSupplier(String methodName) {
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
      synchronized (AgentGroupServiceGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new AgentGroupServiceFileDescriptorSupplier())
              .addMethod(getCreateGroupMethod())
              .addMethod(getEditGroupMethod())
              .addMethod(getListGroupsMethod())
              .addMethod(getDeleteGroupMethod())
              .build();
        }
      }
    }
    return result;
  }
}
