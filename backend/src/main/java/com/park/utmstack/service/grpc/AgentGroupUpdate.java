// Generated by the protocol buffer compiler.  DO NOT EDIT!
// source: agent.proto

package com.park.utmstack.service.grpc;

/**
 * Protobuf type {@code agent.AgentGroupUpdate}
 */
public final class AgentGroupUpdate extends
    com.google.protobuf.GeneratedMessageV3 implements
    // @@protoc_insertion_point(message_implements:agent.AgentGroupUpdate)
    AgentGroupUpdateOrBuilder {
private static final long serialVersionUID = 0L;
  // Use AgentGroupUpdate.newBuilder() to construct.
  private AgentGroupUpdate(com.google.protobuf.GeneratedMessageV3.Builder<?> builder) {
    super(builder);
  }
  private AgentGroupUpdate() {
  }

  @java.lang.Override
  @SuppressWarnings({"unused"})
  protected java.lang.Object newInstance(
      UnusedPrivateParameter unused) {
    return new AgentGroupUpdate();
  }

  public static final com.google.protobuf.Descriptors.Descriptor
      getDescriptor() {
    return com.park.utmstack.service.grpc.AgentManagerGrpc.internal_static_agent_AgentGroupUpdate_descriptor;
  }

  @java.lang.Override
  protected com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
      internalGetFieldAccessorTable() {
    return com.park.utmstack.service.grpc.AgentManagerGrpc.internal_static_agent_AgentGroupUpdate_fieldAccessorTable
        .ensureFieldAccessorsInitialized(
            com.park.utmstack.service.grpc.AgentGroupUpdate.class, com.park.utmstack.service.grpc.AgentGroupUpdate.Builder.class);
  }

  public static final int AGENT_ID_FIELD_NUMBER = 1;
  private int agentId_ = 0;
  /**
   * <code>uint32 agent_id = 1;</code>
   * @return The agentId.
   */
  @java.lang.Override
  public int getAgentId() {
    return agentId_;
  }

  public static final int AGENT_GROUP_FIELD_NUMBER = 3;
  private int agentGroup_ = 0;
  /**
   * <code>uint32 agent_group = 3;</code>
   * @return The agentGroup.
   */
  @java.lang.Override
  public int getAgentGroup() {
    return agentGroup_;
  }

  private byte memoizedIsInitialized = -1;
  @java.lang.Override
  public final boolean isInitialized() {
    byte isInitialized = memoizedIsInitialized;
    if (isInitialized == 1) return true;
    if (isInitialized == 0) return false;

    memoizedIsInitialized = 1;
    return true;
  }

  @java.lang.Override
  public void writeTo(com.google.protobuf.CodedOutputStream output)
                      throws java.io.IOException {
    if (agentId_ != 0) {
      output.writeUInt32(1, agentId_);
    }
    if (agentGroup_ != 0) {
      output.writeUInt32(3, agentGroup_);
    }
    getUnknownFields().writeTo(output);
  }

  @java.lang.Override
  public int getSerializedSize() {
    int size = memoizedSize;
    if (size != -1) return size;

    size = 0;
    if (agentId_ != 0) {
      size += com.google.protobuf.CodedOutputStream
        .computeUInt32Size(1, agentId_);
    }
    if (agentGroup_ != 0) {
      size += com.google.protobuf.CodedOutputStream
        .computeUInt32Size(3, agentGroup_);
    }
    size += getUnknownFields().getSerializedSize();
    memoizedSize = size;
    return size;
  }

  @java.lang.Override
  public boolean equals(final java.lang.Object obj) {
    if (obj == this) {
     return true;
    }
    if (!(obj instanceof com.park.utmstack.service.grpc.AgentGroupUpdate)) {
      return super.equals(obj);
    }
    com.park.utmstack.service.grpc.AgentGroupUpdate other = (com.park.utmstack.service.grpc.AgentGroupUpdate) obj;

    if (getAgentId()
        != other.getAgentId()) return false;
    if (getAgentGroup()
        != other.getAgentGroup()) return false;
    if (!getUnknownFields().equals(other.getUnknownFields())) return false;
    return true;
  }

  @java.lang.Override
  public int hashCode() {
    if (memoizedHashCode != 0) {
      return memoizedHashCode;
    }
    int hash = 41;
    hash = (19 * hash) + getDescriptor().hashCode();
    hash = (37 * hash) + AGENT_ID_FIELD_NUMBER;
    hash = (53 * hash) + getAgentId();
    hash = (37 * hash) + AGENT_GROUP_FIELD_NUMBER;
    hash = (53 * hash) + getAgentGroup();
    hash = (29 * hash) + getUnknownFields().hashCode();
    memoizedHashCode = hash;
    return hash;
  }

  public static com.park.utmstack.service.grpc.AgentGroupUpdate parseFrom(
      java.nio.ByteBuffer data)
      throws com.google.protobuf.InvalidProtocolBufferException {
    return PARSER.parseFrom(data);
  }
  public static com.park.utmstack.service.grpc.AgentGroupUpdate parseFrom(
      java.nio.ByteBuffer data,
      com.google.protobuf.ExtensionRegistryLite extensionRegistry)
      throws com.google.protobuf.InvalidProtocolBufferException {
    return PARSER.parseFrom(data, extensionRegistry);
  }
  public static com.park.utmstack.service.grpc.AgentGroupUpdate parseFrom(
      com.google.protobuf.ByteString data)
      throws com.google.protobuf.InvalidProtocolBufferException {
    return PARSER.parseFrom(data);
  }
  public static com.park.utmstack.service.grpc.AgentGroupUpdate parseFrom(
      com.google.protobuf.ByteString data,
      com.google.protobuf.ExtensionRegistryLite extensionRegistry)
      throws com.google.protobuf.InvalidProtocolBufferException {
    return PARSER.parseFrom(data, extensionRegistry);
  }
  public static com.park.utmstack.service.grpc.AgentGroupUpdate parseFrom(byte[] data)
      throws com.google.protobuf.InvalidProtocolBufferException {
    return PARSER.parseFrom(data);
  }
  public static com.park.utmstack.service.grpc.AgentGroupUpdate parseFrom(
      byte[] data,
      com.google.protobuf.ExtensionRegistryLite extensionRegistry)
      throws com.google.protobuf.InvalidProtocolBufferException {
    return PARSER.parseFrom(data, extensionRegistry);
  }
  public static com.park.utmstack.service.grpc.AgentGroupUpdate parseFrom(java.io.InputStream input)
      throws java.io.IOException {
    return com.google.protobuf.GeneratedMessageV3
        .parseWithIOException(PARSER, input);
  }
  public static com.park.utmstack.service.grpc.AgentGroupUpdate parseFrom(
      java.io.InputStream input,
      com.google.protobuf.ExtensionRegistryLite extensionRegistry)
      throws java.io.IOException {
    return com.google.protobuf.GeneratedMessageV3
        .parseWithIOException(PARSER, input, extensionRegistry);
  }

  public static com.park.utmstack.service.grpc.AgentGroupUpdate parseDelimitedFrom(java.io.InputStream input)
      throws java.io.IOException {
    return com.google.protobuf.GeneratedMessageV3
        .parseDelimitedWithIOException(PARSER, input);
  }

  public static com.park.utmstack.service.grpc.AgentGroupUpdate parseDelimitedFrom(
      java.io.InputStream input,
      com.google.protobuf.ExtensionRegistryLite extensionRegistry)
      throws java.io.IOException {
    return com.google.protobuf.GeneratedMessageV3
        .parseDelimitedWithIOException(PARSER, input, extensionRegistry);
  }
  public static com.park.utmstack.service.grpc.AgentGroupUpdate parseFrom(
      com.google.protobuf.CodedInputStream input)
      throws java.io.IOException {
    return com.google.protobuf.GeneratedMessageV3
        .parseWithIOException(PARSER, input);
  }
  public static com.park.utmstack.service.grpc.AgentGroupUpdate parseFrom(
      com.google.protobuf.CodedInputStream input,
      com.google.protobuf.ExtensionRegistryLite extensionRegistry)
      throws java.io.IOException {
    return com.google.protobuf.GeneratedMessageV3
        .parseWithIOException(PARSER, input, extensionRegistry);
  }

  @java.lang.Override
  public Builder newBuilderForType() { return newBuilder(); }
  public static Builder newBuilder() {
    return DEFAULT_INSTANCE.toBuilder();
  }
  public static Builder newBuilder(com.park.utmstack.service.grpc.AgentGroupUpdate prototype) {
    return DEFAULT_INSTANCE.toBuilder().mergeFrom(prototype);
  }
  @java.lang.Override
  public Builder toBuilder() {
    return this == DEFAULT_INSTANCE
        ? new Builder() : new Builder().mergeFrom(this);
  }

  @java.lang.Override
  protected Builder newBuilderForType(
      com.google.protobuf.GeneratedMessageV3.BuilderParent parent) {
    Builder builder = new Builder(parent);
    return builder;
  }
  /**
   * Protobuf type {@code agent.AgentGroupUpdate}
   */
  public static final class Builder extends
      com.google.protobuf.GeneratedMessageV3.Builder<Builder> implements
      // @@protoc_insertion_point(builder_implements:agent.AgentGroupUpdate)
      com.park.utmstack.service.grpc.AgentGroupUpdateOrBuilder {
    public static final com.google.protobuf.Descriptors.Descriptor
        getDescriptor() {
      return com.park.utmstack.service.grpc.AgentManagerGrpc.internal_static_agent_AgentGroupUpdate_descriptor;
    }

    @java.lang.Override
    protected com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
        internalGetFieldAccessorTable() {
      return com.park.utmstack.service.grpc.AgentManagerGrpc.internal_static_agent_AgentGroupUpdate_fieldAccessorTable
          .ensureFieldAccessorsInitialized(
              com.park.utmstack.service.grpc.AgentGroupUpdate.class, com.park.utmstack.service.grpc.AgentGroupUpdate.Builder.class);
    }

    // Construct using com.park.utmstack.service.grpc.AgentGroupUpdate.newBuilder()
    private Builder() {

    }

    private Builder(
        com.google.protobuf.GeneratedMessageV3.BuilderParent parent) {
      super(parent);

    }
    @java.lang.Override
    public Builder clear() {
      super.clear();
      bitField0_ = 0;
      agentId_ = 0;
      agentGroup_ = 0;
      return this;
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.Descriptor
        getDescriptorForType() {
      return com.park.utmstack.service.grpc.AgentManagerGrpc.internal_static_agent_AgentGroupUpdate_descriptor;
    }

    @java.lang.Override
    public com.park.utmstack.service.grpc.AgentGroupUpdate getDefaultInstanceForType() {
      return com.park.utmstack.service.grpc.AgentGroupUpdate.getDefaultInstance();
    }

    @java.lang.Override
    public com.park.utmstack.service.grpc.AgentGroupUpdate build() {
      com.park.utmstack.service.grpc.AgentGroupUpdate result = buildPartial();
      if (!result.isInitialized()) {
        throw newUninitializedMessageException(result);
      }
      return result;
    }

    @java.lang.Override
    public com.park.utmstack.service.grpc.AgentGroupUpdate buildPartial() {
      com.park.utmstack.service.grpc.AgentGroupUpdate result = new com.park.utmstack.service.grpc.AgentGroupUpdate(this);
      if (bitField0_ != 0) { buildPartial0(result); }
      onBuilt();
      return result;
    }

    private void buildPartial0(com.park.utmstack.service.grpc.AgentGroupUpdate result) {
      int from_bitField0_ = bitField0_;
      if (((from_bitField0_ & 0x00000001) != 0)) {
        result.agentId_ = agentId_;
      }
      if (((from_bitField0_ & 0x00000002) != 0)) {
        result.agentGroup_ = agentGroup_;
      }
    }

    @java.lang.Override
    public Builder clone() {
      return super.clone();
    }
    @java.lang.Override
    public Builder setField(
        com.google.protobuf.Descriptors.FieldDescriptor field,
        java.lang.Object value) {
      return super.setField(field, value);
    }
    @java.lang.Override
    public Builder clearField(
        com.google.protobuf.Descriptors.FieldDescriptor field) {
      return super.clearField(field);
    }
    @java.lang.Override
    public Builder clearOneof(
        com.google.protobuf.Descriptors.OneofDescriptor oneof) {
      return super.clearOneof(oneof);
    }
    @java.lang.Override
    public Builder setRepeatedField(
        com.google.protobuf.Descriptors.FieldDescriptor field,
        int index, java.lang.Object value) {
      return super.setRepeatedField(field, index, value);
    }
    @java.lang.Override
    public Builder addRepeatedField(
        com.google.protobuf.Descriptors.FieldDescriptor field,
        java.lang.Object value) {
      return super.addRepeatedField(field, value);
    }
    @java.lang.Override
    public Builder mergeFrom(com.google.protobuf.Message other) {
      if (other instanceof com.park.utmstack.service.grpc.AgentGroupUpdate) {
        return mergeFrom((com.park.utmstack.service.grpc.AgentGroupUpdate)other);
      } else {
        super.mergeFrom(other);
        return this;
      }
    }

    public Builder mergeFrom(com.park.utmstack.service.grpc.AgentGroupUpdate other) {
      if (other == com.park.utmstack.service.grpc.AgentGroupUpdate.getDefaultInstance()) return this;
      if (other.getAgentId() != 0) {
        setAgentId(other.getAgentId());
      }
      if (other.getAgentGroup() != 0) {
        setAgentGroup(other.getAgentGroup());
      }
      this.mergeUnknownFields(other.getUnknownFields());
      onChanged();
      return this;
    }

    @java.lang.Override
    public final boolean isInitialized() {
      return true;
    }

    @java.lang.Override
    public Builder mergeFrom(
        com.google.protobuf.CodedInputStream input,
        com.google.protobuf.ExtensionRegistryLite extensionRegistry)
        throws java.io.IOException {
      if (extensionRegistry == null) {
        throw new java.lang.NullPointerException();
      }
      try {
        boolean done = false;
        while (!done) {
          int tag = input.readTag();
          switch (tag) {
            case 0:
              done = true;
              break;
            case 8: {
              agentId_ = input.readUInt32();
              bitField0_ |= 0x00000001;
              break;
            } // case 8
            case 24: {
              agentGroup_ = input.readUInt32();
              bitField0_ |= 0x00000002;
              break;
            } // case 24
            default: {
              if (!super.parseUnknownField(input, extensionRegistry, tag)) {
                done = true; // was an endgroup tag
              }
              break;
            } // default:
          } // switch (tag)
        } // while (!done)
      } catch (com.google.protobuf.InvalidProtocolBufferException e) {
        throw e.unwrapIOException();
      } finally {
        onChanged();
      } // finally
      return this;
    }
    private int bitField0_;

    private int agentId_ ;
    /**
     * <code>uint32 agent_id = 1;</code>
     * @return The agentId.
     */
    @java.lang.Override
    public int getAgentId() {
      return agentId_;
    }
    /**
     * <code>uint32 agent_id = 1;</code>
     * @param value The agentId to set.
     * @return This builder for chaining.
     */
    public Builder setAgentId(int value) {

      agentId_ = value;
      bitField0_ |= 0x00000001;
      onChanged();
      return this;
    }
    /**
     * <code>uint32 agent_id = 1;</code>
     * @return This builder for chaining.
     */
    public Builder clearAgentId() {
      bitField0_ = (bitField0_ & ~0x00000001);
      agentId_ = 0;
      onChanged();
      return this;
    }

    private int agentGroup_ ;
    /**
     * <code>uint32 agent_group = 3;</code>
     * @return The agentGroup.
     */
    @java.lang.Override
    public int getAgentGroup() {
      return agentGroup_;
    }
    /**
     * <code>uint32 agent_group = 3;</code>
     * @param value The agentGroup to set.
     * @return This builder for chaining.
     */
    public Builder setAgentGroup(int value) {

      agentGroup_ = value;
      bitField0_ |= 0x00000002;
      onChanged();
      return this;
    }
    /**
     * <code>uint32 agent_group = 3;</code>
     * @return This builder for chaining.
     */
    public Builder clearAgentGroup() {
      bitField0_ = (bitField0_ & ~0x00000002);
      agentGroup_ = 0;
      onChanged();
      return this;
    }
    @java.lang.Override
    public final Builder setUnknownFields(
        final com.google.protobuf.UnknownFieldSet unknownFields) {
      return super.setUnknownFields(unknownFields);
    }

    @java.lang.Override
    public final Builder mergeUnknownFields(
        final com.google.protobuf.UnknownFieldSet unknownFields) {
      return super.mergeUnknownFields(unknownFields);
    }


    // @@protoc_insertion_point(builder_scope:agent.AgentGroupUpdate)
  }

  // @@protoc_insertion_point(class_scope:agent.AgentGroupUpdate)
  private static final com.park.utmstack.service.grpc.AgentGroupUpdate DEFAULT_INSTANCE;
  static {
    DEFAULT_INSTANCE = new com.park.utmstack.service.grpc.AgentGroupUpdate();
  }

  public static com.park.utmstack.service.grpc.AgentGroupUpdate getDefaultInstance() {
    return DEFAULT_INSTANCE;
  }

  private static final com.google.protobuf.Parser<AgentGroupUpdate>
      PARSER = new com.google.protobuf.AbstractParser<AgentGroupUpdate>() {
    @java.lang.Override
    public AgentGroupUpdate parsePartialFrom(
        com.google.protobuf.CodedInputStream input,
        com.google.protobuf.ExtensionRegistryLite extensionRegistry)
        throws com.google.protobuf.InvalidProtocolBufferException {
      Builder builder = newBuilder();
      try {
        builder.mergeFrom(input, extensionRegistry);
      } catch (com.google.protobuf.InvalidProtocolBufferException e) {
        throw e.setUnfinishedMessage(builder.buildPartial());
      } catch (com.google.protobuf.UninitializedMessageException e) {
        throw e.asInvalidProtocolBufferException().setUnfinishedMessage(builder.buildPartial());
      } catch (java.io.IOException e) {
        throw new com.google.protobuf.InvalidProtocolBufferException(e)
            .setUnfinishedMessage(builder.buildPartial());
      }
      return builder.buildPartial();
    }
  };

  public static com.google.protobuf.Parser<AgentGroupUpdate> parser() {
    return PARSER;
  }

  @java.lang.Override
  public com.google.protobuf.Parser<AgentGroupUpdate> getParserForType() {
    return PARSER;
  }

  @java.lang.Override
  public com.park.utmstack.service.grpc.AgentGroupUpdate getDefaultInstanceForType() {
    return DEFAULT_INSTANCE;
  }

}

