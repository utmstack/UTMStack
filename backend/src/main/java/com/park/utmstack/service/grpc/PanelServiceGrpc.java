package com.park.utmstack.service.grpc;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.55.1)",
    comments = "Source: agent.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class PanelServiceGrpc {

  private PanelServiceGrpc() {}

  public static final String SERVICE_NAME = "agent.PanelService";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.UtmCommand,
      com.park.utmstack.service.grpc.CommandResult> getProcessCommandMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ProcessCommand",
      requestType = com.park.utmstack.service.grpc.UtmCommand.class,
      responseType = com.park.utmstack.service.grpc.CommandResult.class,
      methodType = io.grpc.MethodDescriptor.MethodType.BIDI_STREAMING)
  public static io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.UtmCommand,
      com.park.utmstack.service.grpc.CommandResult> getProcessCommandMethod() {
    io.grpc.MethodDescriptor<com.park.utmstack.service.grpc.UtmCommand, com.park.utmstack.service.grpc.CommandResult> getProcessCommandMethod;
    if ((getProcessCommandMethod = PanelServiceGrpc.getProcessCommandMethod) == null) {
      synchronized (PanelServiceGrpc.class) {
        if ((getProcessCommandMethod = PanelServiceGrpc.getProcessCommandMethod) == null) {
          PanelServiceGrpc.getProcessCommandMethod = getProcessCommandMethod =
              io.grpc.MethodDescriptor.<com.park.utmstack.service.grpc.UtmCommand, com.park.utmstack.service.grpc.CommandResult>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.BIDI_STREAMING)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ProcessCommand"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.park.utmstack.service.grpc.UtmCommand.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.park.utmstack.service.grpc.CommandResult.getDefaultInstance()))
              .setSchemaDescriptor(new PanelServiceMethodDescriptorSupplier("ProcessCommand"))
              .build();
        }
      }
    }
    return getProcessCommandMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static PanelServiceStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<PanelServiceStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<PanelServiceStub>() {
        @java.lang.Override
        public PanelServiceStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new PanelServiceStub(channel, callOptions);
        }
      };
    return PanelServiceStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static PanelServiceBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<PanelServiceBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<PanelServiceBlockingStub>() {
        @java.lang.Override
        public PanelServiceBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new PanelServiceBlockingStub(channel, callOptions);
        }
      };
    return PanelServiceBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static PanelServiceFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<PanelServiceFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<PanelServiceFutureStub>() {
        @java.lang.Override
        public PanelServiceFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new PanelServiceFutureStub(channel, callOptions);
        }
      };
    return PanelServiceFutureStub.newStub(factory, channel);
  }

  /**
   */
  public interface AsyncService {

    /**
     */
    default io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.UtmCommand> processCommand(
        io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.CommandResult> responseObserver) {
      return io.grpc.stub.ServerCalls.asyncUnimplementedStreamingCall(getProcessCommandMethod(), responseObserver);
    }
  }

  /**
   * Base class for the server implementation of the service PanelService.
   */
  public static abstract class PanelServiceImplBase
      implements io.grpc.BindableService, AsyncService {

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return PanelServiceGrpc.bindService(this);
    }
  }

  /**
   * A stub to allow clients to do asynchronous rpc calls to service PanelService.
   */
  public static final class PanelServiceStub
      extends io.grpc.stub.AbstractAsyncStub<PanelServiceStub> {
    private PanelServiceStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected PanelServiceStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new PanelServiceStub(channel, callOptions);
    }

    /**
     */
    public io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.UtmCommand> processCommand(
        io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.CommandResult> responseObserver) {
      return io.grpc.stub.ClientCalls.asyncBidiStreamingCall(
          getChannel().newCall(getProcessCommandMethod(), getCallOptions()), responseObserver);
    }
  }

  /**
   * A stub to allow clients to do synchronous rpc calls to service PanelService.
   */
  public static final class PanelServiceBlockingStub
      extends io.grpc.stub.AbstractBlockingStub<PanelServiceBlockingStub> {
    private PanelServiceBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected PanelServiceBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new PanelServiceBlockingStub(channel, callOptions);
    }
  }

  /**
   * A stub to allow clients to do ListenableFuture-style rpc calls to service PanelService.
   */
  public static final class PanelServiceFutureStub
      extends io.grpc.stub.AbstractFutureStub<PanelServiceFutureStub> {
    private PanelServiceFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected PanelServiceFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new PanelServiceFutureStub(channel, callOptions);
    }
  }

  private static final int METHODID_PROCESS_COMMAND = 0;

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
        default:
          throw new AssertionError();
      }
    }

    @java.lang.Override
    @java.lang.SuppressWarnings("unchecked")
    public io.grpc.stub.StreamObserver<Req> invoke(
        io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        case METHODID_PROCESS_COMMAND:
          return (io.grpc.stub.StreamObserver<Req>) serviceImpl.processCommand(
              (io.grpc.stub.StreamObserver<com.park.utmstack.service.grpc.CommandResult>) responseObserver);
        default:
          throw new AssertionError();
      }
    }
  }

  public static final io.grpc.ServerServiceDefinition bindService(AsyncService service) {
    return io.grpc.ServerServiceDefinition.builder(getServiceDescriptor())
        .addMethod(
          getProcessCommandMethod(),
          io.grpc.stub.ServerCalls.asyncBidiStreamingCall(
            new MethodHandlers<
              com.park.utmstack.service.grpc.UtmCommand,
              com.park.utmstack.service.grpc.CommandResult>(
                service, METHODID_PROCESS_COMMAND)))
        .build();
  }

  private static abstract class PanelServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    PanelServiceBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return com.park.utmstack.service.grpc.AgentManagerGrpc.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("PanelService");
    }
  }

  private static final class PanelServiceFileDescriptorSupplier
      extends PanelServiceBaseDescriptorSupplier {
    PanelServiceFileDescriptorSupplier() {}
  }

  private static final class PanelServiceMethodDescriptorSupplier
      extends PanelServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final String methodName;

    PanelServiceMethodDescriptorSupplier(String methodName) {
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
      synchronized (PanelServiceGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new PanelServiceFileDescriptorSupplier())
              .addMethod(getProcessCommandMethod())
              .build();
        }
      }
    }
    return result;
  }
}
