// Generated by the protocol buffer compiler.  DO NOT EDIT!
// NO CHECKED-IN PROTOBUF GENCODE
// source: agent.proto
// Protobuf Java Version: 4.29.3

package com.park.utmstack.service.grpc;

/**
 * Protobuf service {@code agent.AgentService}
 */
public  abstract class AgentService
    implements com.google.protobuf.Service {
  protected AgentService() {}

  public interface Interface {
    /**
     * <code>rpc UpdateAgent(.agent.AgentRequest) returns (.agent.AuthResponse);</code>
     */
    public abstract void updateAgent(
        com.google.protobuf.RpcController controller,
        com.park.utmstack.service.grpc.AgentRequest request,
        com.google.protobuf.RpcCallback<com.park.utmstack.service.grpc.AuthResponse> done);

    /**
     * <code>rpc RegisterAgent(.agent.AgentRequest) returns (.agent.AuthResponse);</code>
     */
    public abstract void registerAgent(
        com.google.protobuf.RpcController controller,
        com.park.utmstack.service.grpc.AgentRequest request,
        com.google.protobuf.RpcCallback<com.park.utmstack.service.grpc.AuthResponse> done);

    /**
     * <code>rpc DeleteAgent(.agent.DeleteRequest) returns (.agent.AuthResponse);</code>
     */
    public abstract void deleteAgent(
        com.google.protobuf.RpcController controller,
        com.park.utmstack.service.grpc.DeleteRequest request,
        com.google.protobuf.RpcCallback<com.park.utmstack.service.grpc.AuthResponse> done);

    /**
     * <code>rpc ListAgents(.agent.ListRequest) returns (.agent.ListAgentsResponse);</code>
     */
    public abstract void listAgents(
        com.google.protobuf.RpcController controller,
        com.park.utmstack.service.grpc.ListRequest request,
        com.google.protobuf.RpcCallback<com.park.utmstack.service.grpc.ListAgentsResponse> done);

    /**
     * <code>rpc AgentStream(stream .agent.BidirectionalStream) returns (stream .agent.BidirectionalStream);</code>
     */
    public abstract void agentStream(
        com.google.protobuf.RpcController controller,
        com.park.utmstack.service.grpc.BidirectionalStream request,
        com.google.protobuf.RpcCallback<com.park.utmstack.service.grpc.BidirectionalStream> done);

    /**
     * <code>rpc ListAgentCommands(.agent.ListRequest) returns (.agent.ListAgentsCommandsResponse);</code>
     */
    public abstract void listAgentCommands(
        com.google.protobuf.RpcController controller,
        com.park.utmstack.service.grpc.ListRequest request,
        com.google.protobuf.RpcCallback<com.park.utmstack.service.grpc.ListAgentsCommandsResponse> done);

  }

  public static com.google.protobuf.Service newReflectiveService(
      final Interface impl) {
    return new AgentService() {
      @java.lang.Override
      public  void updateAgent(
          com.google.protobuf.RpcController controller,
          com.park.utmstack.service.grpc.AgentRequest request,
          com.google.protobuf.RpcCallback<com.park.utmstack.service.grpc.AuthResponse> done) {
        impl.updateAgent(controller, request, done);
      }

      @java.lang.Override
      public  void registerAgent(
          com.google.protobuf.RpcController controller,
          com.park.utmstack.service.grpc.AgentRequest request,
          com.google.protobuf.RpcCallback<com.park.utmstack.service.grpc.AuthResponse> done) {
        impl.registerAgent(controller, request, done);
      }

      @java.lang.Override
      public  void deleteAgent(
          com.google.protobuf.RpcController controller,
          com.park.utmstack.service.grpc.DeleteRequest request,
          com.google.protobuf.RpcCallback<com.park.utmstack.service.grpc.AuthResponse> done) {
        impl.deleteAgent(controller, request, done);
      }

      @java.lang.Override
      public  void listAgents(
          com.google.protobuf.RpcController controller,
          com.park.utmstack.service.grpc.ListRequest request,
          com.google.protobuf.RpcCallback<com.park.utmstack.service.grpc.ListAgentsResponse> done) {
        impl.listAgents(controller, request, done);
      }

      @java.lang.Override
      public  void agentStream(
          com.google.protobuf.RpcController controller,
          com.park.utmstack.service.grpc.BidirectionalStream request,
          com.google.protobuf.RpcCallback<com.park.utmstack.service.grpc.BidirectionalStream> done) {
        impl.agentStream(controller, request, done);
      }

      @java.lang.Override
      public  void listAgentCommands(
          com.google.protobuf.RpcController controller,
          com.park.utmstack.service.grpc.ListRequest request,
          com.google.protobuf.RpcCallback<com.park.utmstack.service.grpc.ListAgentsCommandsResponse> done) {
        impl.listAgentCommands(controller, request, done);
      }

    };
  }

  public static com.google.protobuf.BlockingService
      newReflectiveBlockingService(final BlockingInterface impl) {
    return new com.google.protobuf.BlockingService() {
      public final com.google.protobuf.Descriptors.ServiceDescriptor
          getDescriptorForType() {
        return getDescriptor();
      }

      public final com.google.protobuf.Message callBlockingMethod(
          com.google.protobuf.Descriptors.MethodDescriptor method,
          com.google.protobuf.RpcController controller,
          com.google.protobuf.Message request)
          throws com.google.protobuf.ServiceException {
        if (method.getService() != getDescriptor()) {
          throw new java.lang.IllegalArgumentException(
            "Service.callBlockingMethod() given method descriptor for " +
            "wrong service type.");
        }
        switch(method.getIndex()) {
          case 0:
            return impl.updateAgent(controller, (com.park.utmstack.service.grpc.AgentRequest)request);
          case 1:
            return impl.registerAgent(controller, (com.park.utmstack.service.grpc.AgentRequest)request);
          case 2:
            return impl.deleteAgent(controller, (com.park.utmstack.service.grpc.DeleteRequest)request);
          case 3:
            return impl.listAgents(controller, (com.park.utmstack.service.grpc.ListRequest)request);
          case 4:
            return impl.agentStream(controller, (com.park.utmstack.service.grpc.BidirectionalStream)request);
          case 5:
            return impl.listAgentCommands(controller, (com.park.utmstack.service.grpc.ListRequest)request);
          default:
            throw new java.lang.AssertionError("Can't get here.");
        }
      }

      public final com.google.protobuf.Message
          getRequestPrototype(
          com.google.protobuf.Descriptors.MethodDescriptor method) {
        if (method.getService() != getDescriptor()) {
          throw new java.lang.IllegalArgumentException(
            "Service.getRequestPrototype() given method " +
            "descriptor for wrong service type.");
        }
        switch(method.getIndex()) {
          case 0:
            return com.park.utmstack.service.grpc.AgentRequest.getDefaultInstance();
          case 1:
            return com.park.utmstack.service.grpc.AgentRequest.getDefaultInstance();
          case 2:
            return com.park.utmstack.service.grpc.DeleteRequest.getDefaultInstance();
          case 3:
            return com.park.utmstack.service.grpc.ListRequest.getDefaultInstance();
          case 4:
            return com.park.utmstack.service.grpc.BidirectionalStream.getDefaultInstance();
          case 5:
            return com.park.utmstack.service.grpc.ListRequest.getDefaultInstance();
          default:
            throw new java.lang.AssertionError("Can't get here.");
        }
      }

      public final com.google.protobuf.Message
          getResponsePrototype(
          com.google.protobuf.Descriptors.MethodDescriptor method) {
        if (method.getService() != getDescriptor()) {
          throw new java.lang.IllegalArgumentException(
            "Service.getResponsePrototype() given method " +
            "descriptor for wrong service type.");
        }
        switch(method.getIndex()) {
          case 0:
            return com.park.utmstack.service.grpc.AuthResponse.getDefaultInstance();
          case 1:
            return com.park.utmstack.service.grpc.AuthResponse.getDefaultInstance();
          case 2:
            return com.park.utmstack.service.grpc.AuthResponse.getDefaultInstance();
          case 3:
            return com.park.utmstack.service.grpc.ListAgentsResponse.getDefaultInstance();
          case 4:
            return com.park.utmstack.service.grpc.BidirectionalStream.getDefaultInstance();
          case 5:
            return com.park.utmstack.service.grpc.ListAgentsCommandsResponse.getDefaultInstance();
          default:
            throw new java.lang.AssertionError("Can't get here.");
        }
      }

    };
  }

  /**
   * <code>rpc UpdateAgent(.agent.AgentRequest) returns (.agent.AuthResponse);</code>
   */
  public abstract void updateAgent(
      com.google.protobuf.RpcController controller,
      com.park.utmstack.service.grpc.AgentRequest request,
      com.google.protobuf.RpcCallback<com.park.utmstack.service.grpc.AuthResponse> done);

  /**
   * <code>rpc RegisterAgent(.agent.AgentRequest) returns (.agent.AuthResponse);</code>
   */
  public abstract void registerAgent(
      com.google.protobuf.RpcController controller,
      com.park.utmstack.service.grpc.AgentRequest request,
      com.google.protobuf.RpcCallback<com.park.utmstack.service.grpc.AuthResponse> done);

  /**
   * <code>rpc DeleteAgent(.agent.DeleteRequest) returns (.agent.AuthResponse);</code>
   */
  public abstract void deleteAgent(
      com.google.protobuf.RpcController controller,
      com.park.utmstack.service.grpc.DeleteRequest request,
      com.google.protobuf.RpcCallback<com.park.utmstack.service.grpc.AuthResponse> done);

  /**
   * <code>rpc ListAgents(.agent.ListRequest) returns (.agent.ListAgentsResponse);</code>
   */
  public abstract void listAgents(
      com.google.protobuf.RpcController controller,
      com.park.utmstack.service.grpc.ListRequest request,
      com.google.protobuf.RpcCallback<com.park.utmstack.service.grpc.ListAgentsResponse> done);

  /**
   * <code>rpc AgentStream(stream .agent.BidirectionalStream) returns (stream .agent.BidirectionalStream);</code>
   */
  public abstract void agentStream(
      com.google.protobuf.RpcController controller,
      com.park.utmstack.service.grpc.BidirectionalStream request,
      com.google.protobuf.RpcCallback<com.park.utmstack.service.grpc.BidirectionalStream> done);

  /**
   * <code>rpc ListAgentCommands(.agent.ListRequest) returns (.agent.ListAgentsCommandsResponse);</code>
   */
  public abstract void listAgentCommands(
      com.google.protobuf.RpcController controller,
      com.park.utmstack.service.grpc.ListRequest request,
      com.google.protobuf.RpcCallback<com.park.utmstack.service.grpc.ListAgentsCommandsResponse> done);

  public static final
      com.google.protobuf.Descriptors.ServiceDescriptor
      getDescriptor() {
    return com.park.utmstack.service.grpc.AgentManagerGrpc.getDescriptor().getServices().get(0);
  }
  public final com.google.protobuf.Descriptors.ServiceDescriptor
      getDescriptorForType() {
    return getDescriptor();
  }

  public final void callMethod(
      com.google.protobuf.Descriptors.MethodDescriptor method,
      com.google.protobuf.RpcController controller,
      com.google.protobuf.Message request,
      com.google.protobuf.RpcCallback<
        com.google.protobuf.Message> done) {
    if (method.getService() != getDescriptor()) {
      throw new java.lang.IllegalArgumentException(
        "Service.callMethod() given method descriptor for wrong " +
        "service type.");
    }
    switch(method.getIndex()) {
      case 0:
        this.updateAgent(controller, (com.park.utmstack.service.grpc.AgentRequest)request,
          com.google.protobuf.RpcUtil.<com.park.utmstack.service.grpc.AuthResponse>specializeCallback(
            done));
        return;
      case 1:
        this.registerAgent(controller, (com.park.utmstack.service.grpc.AgentRequest)request,
          com.google.protobuf.RpcUtil.<com.park.utmstack.service.grpc.AuthResponse>specializeCallback(
            done));
        return;
      case 2:
        this.deleteAgent(controller, (com.park.utmstack.service.grpc.DeleteRequest)request,
          com.google.protobuf.RpcUtil.<com.park.utmstack.service.grpc.AuthResponse>specializeCallback(
            done));
        return;
      case 3:
        this.listAgents(controller, (com.park.utmstack.service.grpc.ListRequest)request,
          com.google.protobuf.RpcUtil.<com.park.utmstack.service.grpc.ListAgentsResponse>specializeCallback(
            done));
        return;
      case 4:
        this.agentStream(controller, (com.park.utmstack.service.grpc.BidirectionalStream)request,
          com.google.protobuf.RpcUtil.<com.park.utmstack.service.grpc.BidirectionalStream>specializeCallback(
            done));
        return;
      case 5:
        this.listAgentCommands(controller, (com.park.utmstack.service.grpc.ListRequest)request,
          com.google.protobuf.RpcUtil.<com.park.utmstack.service.grpc.ListAgentsCommandsResponse>specializeCallback(
            done));
        return;
      default:
        throw new java.lang.AssertionError("Can't get here.");
    }
  }

  public final com.google.protobuf.Message
      getRequestPrototype(
      com.google.protobuf.Descriptors.MethodDescriptor method) {
    if (method.getService() != getDescriptor()) {
      throw new java.lang.IllegalArgumentException(
        "Service.getRequestPrototype() given method " +
        "descriptor for wrong service type.");
    }
    switch(method.getIndex()) {
      case 0:
        return com.park.utmstack.service.grpc.AgentRequest.getDefaultInstance();
      case 1:
        return com.park.utmstack.service.grpc.AgentRequest.getDefaultInstance();
      case 2:
        return com.park.utmstack.service.grpc.DeleteRequest.getDefaultInstance();
      case 3:
        return com.park.utmstack.service.grpc.ListRequest.getDefaultInstance();
      case 4:
        return com.park.utmstack.service.grpc.BidirectionalStream.getDefaultInstance();
      case 5:
        return com.park.utmstack.service.grpc.ListRequest.getDefaultInstance();
      default:
        throw new java.lang.AssertionError("Can't get here.");
    }
  }

  public final com.google.protobuf.Message
      getResponsePrototype(
      com.google.protobuf.Descriptors.MethodDescriptor method) {
    if (method.getService() != getDescriptor()) {
      throw new java.lang.IllegalArgumentException(
        "Service.getResponsePrototype() given method " +
        "descriptor for wrong service type.");
    }
    switch(method.getIndex()) {
      case 0:
        return com.park.utmstack.service.grpc.AuthResponse.getDefaultInstance();
      case 1:
        return com.park.utmstack.service.grpc.AuthResponse.getDefaultInstance();
      case 2:
        return com.park.utmstack.service.grpc.AuthResponse.getDefaultInstance();
      case 3:
        return com.park.utmstack.service.grpc.ListAgentsResponse.getDefaultInstance();
      case 4:
        return com.park.utmstack.service.grpc.BidirectionalStream.getDefaultInstance();
      case 5:
        return com.park.utmstack.service.grpc.ListAgentsCommandsResponse.getDefaultInstance();
      default:
        throw new java.lang.AssertionError("Can't get here.");
    }
  }

  public static Stub newStub(
      com.google.protobuf.RpcChannel channel) {
    return new Stub(channel);
  }

  public static final class Stub extends com.park.utmstack.service.grpc.AgentService implements Interface {
    private Stub(com.google.protobuf.RpcChannel channel) {
      this.channel = channel;
    }

    private final com.google.protobuf.RpcChannel channel;

    public com.google.protobuf.RpcChannel getChannel() {
      return channel;
    }

    public  void updateAgent(
        com.google.protobuf.RpcController controller,
        com.park.utmstack.service.grpc.AgentRequest request,
        com.google.protobuf.RpcCallback<com.park.utmstack.service.grpc.AuthResponse> done) {
      channel.callMethod(
        getDescriptor().getMethods().get(0),
        controller,
        request,
        com.park.utmstack.service.grpc.AuthResponse.getDefaultInstance(),
        com.google.protobuf.RpcUtil.generalizeCallback(
          done,
          com.park.utmstack.service.grpc.AuthResponse.class,
          com.park.utmstack.service.grpc.AuthResponse.getDefaultInstance()));
    }

    public  void registerAgent(
        com.google.protobuf.RpcController controller,
        com.park.utmstack.service.grpc.AgentRequest request,
        com.google.protobuf.RpcCallback<com.park.utmstack.service.grpc.AuthResponse> done) {
      channel.callMethod(
        getDescriptor().getMethods().get(1),
        controller,
        request,
        com.park.utmstack.service.grpc.AuthResponse.getDefaultInstance(),
        com.google.protobuf.RpcUtil.generalizeCallback(
          done,
          com.park.utmstack.service.grpc.AuthResponse.class,
          com.park.utmstack.service.grpc.AuthResponse.getDefaultInstance()));
    }

    public  void deleteAgent(
        com.google.protobuf.RpcController controller,
        com.park.utmstack.service.grpc.DeleteRequest request,
        com.google.protobuf.RpcCallback<com.park.utmstack.service.grpc.AuthResponse> done) {
      channel.callMethod(
        getDescriptor().getMethods().get(2),
        controller,
        request,
        com.park.utmstack.service.grpc.AuthResponse.getDefaultInstance(),
        com.google.protobuf.RpcUtil.generalizeCallback(
          done,
          com.park.utmstack.service.grpc.AuthResponse.class,
          com.park.utmstack.service.grpc.AuthResponse.getDefaultInstance()));
    }

    public  void listAgents(
        com.google.protobuf.RpcController controller,
        com.park.utmstack.service.grpc.ListRequest request,
        com.google.protobuf.RpcCallback<com.park.utmstack.service.grpc.ListAgentsResponse> done) {
      channel.callMethod(
        getDescriptor().getMethods().get(3),
        controller,
        request,
        com.park.utmstack.service.grpc.ListAgentsResponse.getDefaultInstance(),
        com.google.protobuf.RpcUtil.generalizeCallback(
          done,
          com.park.utmstack.service.grpc.ListAgentsResponse.class,
          com.park.utmstack.service.grpc.ListAgentsResponse.getDefaultInstance()));
    }

    public  void agentStream(
        com.google.protobuf.RpcController controller,
        com.park.utmstack.service.grpc.BidirectionalStream request,
        com.google.protobuf.RpcCallback<com.park.utmstack.service.grpc.BidirectionalStream> done) {
      channel.callMethod(
        getDescriptor().getMethods().get(4),
        controller,
        request,
        com.park.utmstack.service.grpc.BidirectionalStream.getDefaultInstance(),
        com.google.protobuf.RpcUtil.generalizeCallback(
          done,
          com.park.utmstack.service.grpc.BidirectionalStream.class,
          com.park.utmstack.service.grpc.BidirectionalStream.getDefaultInstance()));
    }

    public  void listAgentCommands(
        com.google.protobuf.RpcController controller,
        com.park.utmstack.service.grpc.ListRequest request,
        com.google.protobuf.RpcCallback<com.park.utmstack.service.grpc.ListAgentsCommandsResponse> done) {
      channel.callMethod(
        getDescriptor().getMethods().get(5),
        controller,
        request,
        com.park.utmstack.service.grpc.ListAgentsCommandsResponse.getDefaultInstance(),
        com.google.protobuf.RpcUtil.generalizeCallback(
          done,
          com.park.utmstack.service.grpc.ListAgentsCommandsResponse.class,
          com.park.utmstack.service.grpc.ListAgentsCommandsResponse.getDefaultInstance()));
    }
  }

  public static BlockingInterface newBlockingStub(
      com.google.protobuf.BlockingRpcChannel channel) {
    return new BlockingStub(channel);
  }

  public interface BlockingInterface {
    public com.park.utmstack.service.grpc.AuthResponse updateAgent(
        com.google.protobuf.RpcController controller,
        com.park.utmstack.service.grpc.AgentRequest request)
        throws com.google.protobuf.ServiceException;

    public com.park.utmstack.service.grpc.AuthResponse registerAgent(
        com.google.protobuf.RpcController controller,
        com.park.utmstack.service.grpc.AgentRequest request)
        throws com.google.protobuf.ServiceException;

    public com.park.utmstack.service.grpc.AuthResponse deleteAgent(
        com.google.protobuf.RpcController controller,
        com.park.utmstack.service.grpc.DeleteRequest request)
        throws com.google.protobuf.ServiceException;

    public com.park.utmstack.service.grpc.ListAgentsResponse listAgents(
        com.google.protobuf.RpcController controller,
        com.park.utmstack.service.grpc.ListRequest request)
        throws com.google.protobuf.ServiceException;

    public com.park.utmstack.service.grpc.BidirectionalStream agentStream(
        com.google.protobuf.RpcController controller,
        com.park.utmstack.service.grpc.BidirectionalStream request)
        throws com.google.protobuf.ServiceException;

    public com.park.utmstack.service.grpc.ListAgentsCommandsResponse listAgentCommands(
        com.google.protobuf.RpcController controller,
        com.park.utmstack.service.grpc.ListRequest request)
        throws com.google.protobuf.ServiceException;
  }

  private static final class BlockingStub implements BlockingInterface {
    private BlockingStub(com.google.protobuf.BlockingRpcChannel channel) {
      this.channel = channel;
    }

    private final com.google.protobuf.BlockingRpcChannel channel;

    public com.park.utmstack.service.grpc.AuthResponse updateAgent(
        com.google.protobuf.RpcController controller,
        com.park.utmstack.service.grpc.AgentRequest request)
        throws com.google.protobuf.ServiceException {
      return (com.park.utmstack.service.grpc.AuthResponse) channel.callBlockingMethod(
        getDescriptor().getMethods().get(0),
        controller,
        request,
        com.park.utmstack.service.grpc.AuthResponse.getDefaultInstance());
    }


    public com.park.utmstack.service.grpc.AuthResponse registerAgent(
        com.google.protobuf.RpcController controller,
        com.park.utmstack.service.grpc.AgentRequest request)
        throws com.google.protobuf.ServiceException {
      return (com.park.utmstack.service.grpc.AuthResponse) channel.callBlockingMethod(
        getDescriptor().getMethods().get(1),
        controller,
        request,
        com.park.utmstack.service.grpc.AuthResponse.getDefaultInstance());
    }


    public com.park.utmstack.service.grpc.AuthResponse deleteAgent(
        com.google.protobuf.RpcController controller,
        com.park.utmstack.service.grpc.DeleteRequest request)
        throws com.google.protobuf.ServiceException {
      return (com.park.utmstack.service.grpc.AuthResponse) channel.callBlockingMethod(
        getDescriptor().getMethods().get(2),
        controller,
        request,
        com.park.utmstack.service.grpc.AuthResponse.getDefaultInstance());
    }


    public com.park.utmstack.service.grpc.ListAgentsResponse listAgents(
        com.google.protobuf.RpcController controller,
        com.park.utmstack.service.grpc.ListRequest request)
        throws com.google.protobuf.ServiceException {
      return (com.park.utmstack.service.grpc.ListAgentsResponse) channel.callBlockingMethod(
        getDescriptor().getMethods().get(3),
        controller,
        request,
        com.park.utmstack.service.grpc.ListAgentsResponse.getDefaultInstance());
    }


    public com.park.utmstack.service.grpc.BidirectionalStream agentStream(
        com.google.protobuf.RpcController controller,
        com.park.utmstack.service.grpc.BidirectionalStream request)
        throws com.google.protobuf.ServiceException {
      return (com.park.utmstack.service.grpc.BidirectionalStream) channel.callBlockingMethod(
        getDescriptor().getMethods().get(4),
        controller,
        request,
        com.park.utmstack.service.grpc.BidirectionalStream.getDefaultInstance());
    }


    public com.park.utmstack.service.grpc.ListAgentsCommandsResponse listAgentCommands(
        com.google.protobuf.RpcController controller,
        com.park.utmstack.service.grpc.ListRequest request)
        throws com.google.protobuf.ServiceException {
      return (com.park.utmstack.service.grpc.ListAgentsCommandsResponse) channel.callBlockingMethod(
        getDescriptor().getMethods().get(5),
        controller,
        request,
        com.park.utmstack.service.grpc.ListAgentsCommandsResponse.getDefaultInstance());
    }

  }

  // @@protoc_insertion_point(class_scope:agent.AgentService)
}

