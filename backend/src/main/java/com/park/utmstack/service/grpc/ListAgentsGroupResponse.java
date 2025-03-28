// Generated by the protocol buffer compiler.  DO NOT EDIT!
// NO CHECKED-IN PROTOBUF GENCODE
// source: agent.proto
// Protobuf Java Version: 4.29.3

package com.park.utmstack.service.grpc;

/**
 * Protobuf type {@code agent.ListAgentsGroupResponse}
 */
public final class ListAgentsGroupResponse extends
    com.google.protobuf.GeneratedMessage implements
    // @@protoc_insertion_point(message_implements:agent.ListAgentsGroupResponse)
    ListAgentsGroupResponseOrBuilder {
private static final long serialVersionUID = 0L;
  static {
    com.google.protobuf.RuntimeVersion.validateProtobufGencodeVersion(
      com.google.protobuf.RuntimeVersion.RuntimeDomain.PUBLIC,
      /* major= */ 4,
      /* minor= */ 29,
      /* patch= */ 3,
      /* suffix= */ "",
      ListAgentsGroupResponse.class.getName());
  }
  // Use ListAgentsGroupResponse.newBuilder() to construct.
  private ListAgentsGroupResponse(com.google.protobuf.GeneratedMessage.Builder<?> builder) {
    super(builder);
  }
  private ListAgentsGroupResponse() {
    rows_ = java.util.Collections.emptyList();
  }

  public static final com.google.protobuf.Descriptors.Descriptor
      getDescriptor() {
    return com.park.utmstack.service.grpc.AgentManagerGrpc.internal_static_agent_ListAgentsGroupResponse_descriptor;
  }

  @java.lang.Override
  protected com.google.protobuf.GeneratedMessage.FieldAccessorTable
      internalGetFieldAccessorTable() {
    return com.park.utmstack.service.grpc.AgentManagerGrpc.internal_static_agent_ListAgentsGroupResponse_fieldAccessorTable
        .ensureFieldAccessorsInitialized(
            com.park.utmstack.service.grpc.ListAgentsGroupResponse.class, com.park.utmstack.service.grpc.ListAgentsGroupResponse.Builder.class);
  }

  public static final int ROWS_FIELD_NUMBER = 1;
  @SuppressWarnings("serial")
  private java.util.List<com.park.utmstack.service.grpc.AgentGroup> rows_;
  /**
   * <code>repeated .agent.AgentGroup rows = 1;</code>
   */
  @java.lang.Override
  public java.util.List<com.park.utmstack.service.grpc.AgentGroup> getRowsList() {
    return rows_;
  }
  /**
   * <code>repeated .agent.AgentGroup rows = 1;</code>
   */
  @java.lang.Override
  public java.util.List<? extends com.park.utmstack.service.grpc.AgentGroupOrBuilder> 
      getRowsOrBuilderList() {
    return rows_;
  }
  /**
   * <code>repeated .agent.AgentGroup rows = 1;</code>
   */
  @java.lang.Override
  public int getRowsCount() {
    return rows_.size();
  }
  /**
   * <code>repeated .agent.AgentGroup rows = 1;</code>
   */
  @java.lang.Override
  public com.park.utmstack.service.grpc.AgentGroup getRows(int index) {
    return rows_.get(index);
  }
  /**
   * <code>repeated .agent.AgentGroup rows = 1;</code>
   */
  @java.lang.Override
  public com.park.utmstack.service.grpc.AgentGroupOrBuilder getRowsOrBuilder(
      int index) {
    return rows_.get(index);
  }

  public static final int TOTAL_FIELD_NUMBER = 2;
  private int total_ = 0;
  /**
   * <code>int32 total = 2;</code>
   * @return The total.
   */
  @java.lang.Override
  public int getTotal() {
    return total_;
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
    for (int i = 0; i < rows_.size(); i++) {
      output.writeMessage(1, rows_.get(i));
    }
    if (total_ != 0) {
      output.writeInt32(2, total_);
    }
    getUnknownFields().writeTo(output);
  }

  @java.lang.Override
  public int getSerializedSize() {
    int size = memoizedSize;
    if (size != -1) return size;

    size = 0;
    for (int i = 0; i < rows_.size(); i++) {
      size += com.google.protobuf.CodedOutputStream
        .computeMessageSize(1, rows_.get(i));
    }
    if (total_ != 0) {
      size += com.google.protobuf.CodedOutputStream
        .computeInt32Size(2, total_);
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
    if (!(obj instanceof com.park.utmstack.service.grpc.ListAgentsGroupResponse)) {
      return super.equals(obj);
    }
    com.park.utmstack.service.grpc.ListAgentsGroupResponse other = (com.park.utmstack.service.grpc.ListAgentsGroupResponse) obj;

    if (!getRowsList()
        .equals(other.getRowsList())) return false;
    if (getTotal()
        != other.getTotal()) return false;
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
    if (getRowsCount() > 0) {
      hash = (37 * hash) + ROWS_FIELD_NUMBER;
      hash = (53 * hash) + getRowsList().hashCode();
    }
    hash = (37 * hash) + TOTAL_FIELD_NUMBER;
    hash = (53 * hash) + getTotal();
    hash = (29 * hash) + getUnknownFields().hashCode();
    memoizedHashCode = hash;
    return hash;
  }

  public static com.park.utmstack.service.grpc.ListAgentsGroupResponse parseFrom(
      java.nio.ByteBuffer data)
      throws com.google.protobuf.InvalidProtocolBufferException {
    return PARSER.parseFrom(data);
  }
  public static com.park.utmstack.service.grpc.ListAgentsGroupResponse parseFrom(
      java.nio.ByteBuffer data,
      com.google.protobuf.ExtensionRegistryLite extensionRegistry)
      throws com.google.protobuf.InvalidProtocolBufferException {
    return PARSER.parseFrom(data, extensionRegistry);
  }
  public static com.park.utmstack.service.grpc.ListAgentsGroupResponse parseFrom(
      com.google.protobuf.ByteString data)
      throws com.google.protobuf.InvalidProtocolBufferException {
    return PARSER.parseFrom(data);
  }
  public static com.park.utmstack.service.grpc.ListAgentsGroupResponse parseFrom(
      com.google.protobuf.ByteString data,
      com.google.protobuf.ExtensionRegistryLite extensionRegistry)
      throws com.google.protobuf.InvalidProtocolBufferException {
    return PARSER.parseFrom(data, extensionRegistry);
  }
  public static com.park.utmstack.service.grpc.ListAgentsGroupResponse parseFrom(byte[] data)
      throws com.google.protobuf.InvalidProtocolBufferException {
    return PARSER.parseFrom(data);
  }
  public static com.park.utmstack.service.grpc.ListAgentsGroupResponse parseFrom(
      byte[] data,
      com.google.protobuf.ExtensionRegistryLite extensionRegistry)
      throws com.google.protobuf.InvalidProtocolBufferException {
    return PARSER.parseFrom(data, extensionRegistry);
  }
  public static com.park.utmstack.service.grpc.ListAgentsGroupResponse parseFrom(java.io.InputStream input)
      throws java.io.IOException {
    return com.google.protobuf.GeneratedMessage
        .parseWithIOException(PARSER, input);
  }
  public static com.park.utmstack.service.grpc.ListAgentsGroupResponse parseFrom(
      java.io.InputStream input,
      com.google.protobuf.ExtensionRegistryLite extensionRegistry)
      throws java.io.IOException {
    return com.google.protobuf.GeneratedMessage
        .parseWithIOException(PARSER, input, extensionRegistry);
  }

  public static com.park.utmstack.service.grpc.ListAgentsGroupResponse parseDelimitedFrom(java.io.InputStream input)
      throws java.io.IOException {
    return com.google.protobuf.GeneratedMessage
        .parseDelimitedWithIOException(PARSER, input);
  }

  public static com.park.utmstack.service.grpc.ListAgentsGroupResponse parseDelimitedFrom(
      java.io.InputStream input,
      com.google.protobuf.ExtensionRegistryLite extensionRegistry)
      throws java.io.IOException {
    return com.google.protobuf.GeneratedMessage
        .parseDelimitedWithIOException(PARSER, input, extensionRegistry);
  }
  public static com.park.utmstack.service.grpc.ListAgentsGroupResponse parseFrom(
      com.google.protobuf.CodedInputStream input)
      throws java.io.IOException {
    return com.google.protobuf.GeneratedMessage
        .parseWithIOException(PARSER, input);
  }
  public static com.park.utmstack.service.grpc.ListAgentsGroupResponse parseFrom(
      com.google.protobuf.CodedInputStream input,
      com.google.protobuf.ExtensionRegistryLite extensionRegistry)
      throws java.io.IOException {
    return com.google.protobuf.GeneratedMessage
        .parseWithIOException(PARSER, input, extensionRegistry);
  }

  @java.lang.Override
  public Builder newBuilderForType() { return newBuilder(); }
  public static Builder newBuilder() {
    return DEFAULT_INSTANCE.toBuilder();
  }
  public static Builder newBuilder(com.park.utmstack.service.grpc.ListAgentsGroupResponse prototype) {
    return DEFAULT_INSTANCE.toBuilder().mergeFrom(prototype);
  }
  @java.lang.Override
  public Builder toBuilder() {
    return this == DEFAULT_INSTANCE
        ? new Builder() : new Builder().mergeFrom(this);
  }

  @java.lang.Override
  protected Builder newBuilderForType(
      com.google.protobuf.GeneratedMessage.BuilderParent parent) {
    Builder builder = new Builder(parent);
    return builder;
  }
  /**
   * Protobuf type {@code agent.ListAgentsGroupResponse}
   */
  public static final class Builder extends
      com.google.protobuf.GeneratedMessage.Builder<Builder> implements
      // @@protoc_insertion_point(builder_implements:agent.ListAgentsGroupResponse)
      com.park.utmstack.service.grpc.ListAgentsGroupResponseOrBuilder {
    public static final com.google.protobuf.Descriptors.Descriptor
        getDescriptor() {
      return com.park.utmstack.service.grpc.AgentManagerGrpc.internal_static_agent_ListAgentsGroupResponse_descriptor;
    }

    @java.lang.Override
    protected com.google.protobuf.GeneratedMessage.FieldAccessorTable
        internalGetFieldAccessorTable() {
      return com.park.utmstack.service.grpc.AgentManagerGrpc.internal_static_agent_ListAgentsGroupResponse_fieldAccessorTable
          .ensureFieldAccessorsInitialized(
              com.park.utmstack.service.grpc.ListAgentsGroupResponse.class, com.park.utmstack.service.grpc.ListAgentsGroupResponse.Builder.class);
    }

    // Construct using com.park.utmstack.service.grpc.ListAgentsGroupResponse.newBuilder()
    private Builder() {

    }

    private Builder(
        com.google.protobuf.GeneratedMessage.BuilderParent parent) {
      super(parent);

    }
    @java.lang.Override
    public Builder clear() {
      super.clear();
      bitField0_ = 0;
      if (rowsBuilder_ == null) {
        rows_ = java.util.Collections.emptyList();
      } else {
        rows_ = null;
        rowsBuilder_.clear();
      }
      bitField0_ = (bitField0_ & ~0x00000001);
      total_ = 0;
      return this;
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.Descriptor
        getDescriptorForType() {
      return com.park.utmstack.service.grpc.AgentManagerGrpc.internal_static_agent_ListAgentsGroupResponse_descriptor;
    }

    @java.lang.Override
    public com.park.utmstack.service.grpc.ListAgentsGroupResponse getDefaultInstanceForType() {
      return com.park.utmstack.service.grpc.ListAgentsGroupResponse.getDefaultInstance();
    }

    @java.lang.Override
    public com.park.utmstack.service.grpc.ListAgentsGroupResponse build() {
      com.park.utmstack.service.grpc.ListAgentsGroupResponse result = buildPartial();
      if (!result.isInitialized()) {
        throw newUninitializedMessageException(result);
      }
      return result;
    }

    @java.lang.Override
    public com.park.utmstack.service.grpc.ListAgentsGroupResponse buildPartial() {
      com.park.utmstack.service.grpc.ListAgentsGroupResponse result = new com.park.utmstack.service.grpc.ListAgentsGroupResponse(this);
      buildPartialRepeatedFields(result);
      if (bitField0_ != 0) { buildPartial0(result); }
      onBuilt();
      return result;
    }

    private void buildPartialRepeatedFields(com.park.utmstack.service.grpc.ListAgentsGroupResponse result) {
      if (rowsBuilder_ == null) {
        if (((bitField0_ & 0x00000001) != 0)) {
          rows_ = java.util.Collections.unmodifiableList(rows_);
          bitField0_ = (bitField0_ & ~0x00000001);
        }
        result.rows_ = rows_;
      } else {
        result.rows_ = rowsBuilder_.build();
      }
    }

    private void buildPartial0(com.park.utmstack.service.grpc.ListAgentsGroupResponse result) {
      int from_bitField0_ = bitField0_;
      if (((from_bitField0_ & 0x00000002) != 0)) {
        result.total_ = total_;
      }
    }

    @java.lang.Override
    public Builder mergeFrom(com.google.protobuf.Message other) {
      if (other instanceof com.park.utmstack.service.grpc.ListAgentsGroupResponse) {
        return mergeFrom((com.park.utmstack.service.grpc.ListAgentsGroupResponse)other);
      } else {
        super.mergeFrom(other);
        return this;
      }
    }

    public Builder mergeFrom(com.park.utmstack.service.grpc.ListAgentsGroupResponse other) {
      if (other == com.park.utmstack.service.grpc.ListAgentsGroupResponse.getDefaultInstance()) return this;
      if (rowsBuilder_ == null) {
        if (!other.rows_.isEmpty()) {
          if (rows_.isEmpty()) {
            rows_ = other.rows_;
            bitField0_ = (bitField0_ & ~0x00000001);
          } else {
            ensureRowsIsMutable();
            rows_.addAll(other.rows_);
          }
          onChanged();
        }
      } else {
        if (!other.rows_.isEmpty()) {
          if (rowsBuilder_.isEmpty()) {
            rowsBuilder_.dispose();
            rowsBuilder_ = null;
            rows_ = other.rows_;
            bitField0_ = (bitField0_ & ~0x00000001);
            rowsBuilder_ = 
              com.google.protobuf.GeneratedMessage.alwaysUseFieldBuilders ?
                 getRowsFieldBuilder() : null;
          } else {
            rowsBuilder_.addAllMessages(other.rows_);
          }
        }
      }
      if (other.getTotal() != 0) {
        setTotal(other.getTotal());
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
            case 10: {
              com.park.utmstack.service.grpc.AgentGroup m =
                  input.readMessage(
                      com.park.utmstack.service.grpc.AgentGroup.parser(),
                      extensionRegistry);
              if (rowsBuilder_ == null) {
                ensureRowsIsMutable();
                rows_.add(m);
              } else {
                rowsBuilder_.addMessage(m);
              }
              break;
            } // case 10
            case 16: {
              total_ = input.readInt32();
              bitField0_ |= 0x00000002;
              break;
            } // case 16
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

    private java.util.List<com.park.utmstack.service.grpc.AgentGroup> rows_ =
      java.util.Collections.emptyList();
    private void ensureRowsIsMutable() {
      if (!((bitField0_ & 0x00000001) != 0)) {
        rows_ = new java.util.ArrayList<com.park.utmstack.service.grpc.AgentGroup>(rows_);
        bitField0_ |= 0x00000001;
       }
    }

    private com.google.protobuf.RepeatedFieldBuilder<
        com.park.utmstack.service.grpc.AgentGroup, com.park.utmstack.service.grpc.AgentGroup.Builder, com.park.utmstack.service.grpc.AgentGroupOrBuilder> rowsBuilder_;

    /**
     * <code>repeated .agent.AgentGroup rows = 1;</code>
     */
    public java.util.List<com.park.utmstack.service.grpc.AgentGroup> getRowsList() {
      if (rowsBuilder_ == null) {
        return java.util.Collections.unmodifiableList(rows_);
      } else {
        return rowsBuilder_.getMessageList();
      }
    }
    /**
     * <code>repeated .agent.AgentGroup rows = 1;</code>
     */
    public int getRowsCount() {
      if (rowsBuilder_ == null) {
        return rows_.size();
      } else {
        return rowsBuilder_.getCount();
      }
    }
    /**
     * <code>repeated .agent.AgentGroup rows = 1;</code>
     */
    public com.park.utmstack.service.grpc.AgentGroup getRows(int index) {
      if (rowsBuilder_ == null) {
        return rows_.get(index);
      } else {
        return rowsBuilder_.getMessage(index);
      }
    }
    /**
     * <code>repeated .agent.AgentGroup rows = 1;</code>
     */
    public Builder setRows(
        int index, com.park.utmstack.service.grpc.AgentGroup value) {
      if (rowsBuilder_ == null) {
        if (value == null) {
          throw new NullPointerException();
        }
        ensureRowsIsMutable();
        rows_.set(index, value);
        onChanged();
      } else {
        rowsBuilder_.setMessage(index, value);
      }
      return this;
    }
    /**
     * <code>repeated .agent.AgentGroup rows = 1;</code>
     */
    public Builder setRows(
        int index, com.park.utmstack.service.grpc.AgentGroup.Builder builderForValue) {
      if (rowsBuilder_ == null) {
        ensureRowsIsMutable();
        rows_.set(index, builderForValue.build());
        onChanged();
      } else {
        rowsBuilder_.setMessage(index, builderForValue.build());
      }
      return this;
    }
    /**
     * <code>repeated .agent.AgentGroup rows = 1;</code>
     */
    public Builder addRows(com.park.utmstack.service.grpc.AgentGroup value) {
      if (rowsBuilder_ == null) {
        if (value == null) {
          throw new NullPointerException();
        }
        ensureRowsIsMutable();
        rows_.add(value);
        onChanged();
      } else {
        rowsBuilder_.addMessage(value);
      }
      return this;
    }
    /**
     * <code>repeated .agent.AgentGroup rows = 1;</code>
     */
    public Builder addRows(
        int index, com.park.utmstack.service.grpc.AgentGroup value) {
      if (rowsBuilder_ == null) {
        if (value == null) {
          throw new NullPointerException();
        }
        ensureRowsIsMutable();
        rows_.add(index, value);
        onChanged();
      } else {
        rowsBuilder_.addMessage(index, value);
      }
      return this;
    }
    /**
     * <code>repeated .agent.AgentGroup rows = 1;</code>
     */
    public Builder addRows(
        com.park.utmstack.service.grpc.AgentGroup.Builder builderForValue) {
      if (rowsBuilder_ == null) {
        ensureRowsIsMutable();
        rows_.add(builderForValue.build());
        onChanged();
      } else {
        rowsBuilder_.addMessage(builderForValue.build());
      }
      return this;
    }
    /**
     * <code>repeated .agent.AgentGroup rows = 1;</code>
     */
    public Builder addRows(
        int index, com.park.utmstack.service.grpc.AgentGroup.Builder builderForValue) {
      if (rowsBuilder_ == null) {
        ensureRowsIsMutable();
        rows_.add(index, builderForValue.build());
        onChanged();
      } else {
        rowsBuilder_.addMessage(index, builderForValue.build());
      }
      return this;
    }
    /**
     * <code>repeated .agent.AgentGroup rows = 1;</code>
     */
    public Builder addAllRows(
        java.lang.Iterable<? extends com.park.utmstack.service.grpc.AgentGroup> values) {
      if (rowsBuilder_ == null) {
        ensureRowsIsMutable();
        com.google.protobuf.AbstractMessageLite.Builder.addAll(
            values, rows_);
        onChanged();
      } else {
        rowsBuilder_.addAllMessages(values);
      }
      return this;
    }
    /**
     * <code>repeated .agent.AgentGroup rows = 1;</code>
     */
    public Builder clearRows() {
      if (rowsBuilder_ == null) {
        rows_ = java.util.Collections.emptyList();
        bitField0_ = (bitField0_ & ~0x00000001);
        onChanged();
      } else {
        rowsBuilder_.clear();
      }
      return this;
    }
    /**
     * <code>repeated .agent.AgentGroup rows = 1;</code>
     */
    public Builder removeRows(int index) {
      if (rowsBuilder_ == null) {
        ensureRowsIsMutable();
        rows_.remove(index);
        onChanged();
      } else {
        rowsBuilder_.remove(index);
      }
      return this;
    }
    /**
     * <code>repeated .agent.AgentGroup rows = 1;</code>
     */
    public com.park.utmstack.service.grpc.AgentGroup.Builder getRowsBuilder(
        int index) {
      return getRowsFieldBuilder().getBuilder(index);
    }
    /**
     * <code>repeated .agent.AgentGroup rows = 1;</code>
     */
    public com.park.utmstack.service.grpc.AgentGroupOrBuilder getRowsOrBuilder(
        int index) {
      if (rowsBuilder_ == null) {
        return rows_.get(index);  } else {
        return rowsBuilder_.getMessageOrBuilder(index);
      }
    }
    /**
     * <code>repeated .agent.AgentGroup rows = 1;</code>
     */
    public java.util.List<? extends com.park.utmstack.service.grpc.AgentGroupOrBuilder> 
         getRowsOrBuilderList() {
      if (rowsBuilder_ != null) {
        return rowsBuilder_.getMessageOrBuilderList();
      } else {
        return java.util.Collections.unmodifiableList(rows_);
      }
    }
    /**
     * <code>repeated .agent.AgentGroup rows = 1;</code>
     */
    public com.park.utmstack.service.grpc.AgentGroup.Builder addRowsBuilder() {
      return getRowsFieldBuilder().addBuilder(
          com.park.utmstack.service.grpc.AgentGroup.getDefaultInstance());
    }
    /**
     * <code>repeated .agent.AgentGroup rows = 1;</code>
     */
    public com.park.utmstack.service.grpc.AgentGroup.Builder addRowsBuilder(
        int index) {
      return getRowsFieldBuilder().addBuilder(
          index, com.park.utmstack.service.grpc.AgentGroup.getDefaultInstance());
    }
    /**
     * <code>repeated .agent.AgentGroup rows = 1;</code>
     */
    public java.util.List<com.park.utmstack.service.grpc.AgentGroup.Builder> 
         getRowsBuilderList() {
      return getRowsFieldBuilder().getBuilderList();
    }
    private com.google.protobuf.RepeatedFieldBuilder<
        com.park.utmstack.service.grpc.AgentGroup, com.park.utmstack.service.grpc.AgentGroup.Builder, com.park.utmstack.service.grpc.AgentGroupOrBuilder> 
        getRowsFieldBuilder() {
      if (rowsBuilder_ == null) {
        rowsBuilder_ = new com.google.protobuf.RepeatedFieldBuilder<
            com.park.utmstack.service.grpc.AgentGroup, com.park.utmstack.service.grpc.AgentGroup.Builder, com.park.utmstack.service.grpc.AgentGroupOrBuilder>(
                rows_,
                ((bitField0_ & 0x00000001) != 0),
                getParentForChildren(),
                isClean());
        rows_ = null;
      }
      return rowsBuilder_;
    }

    private int total_ ;
    /**
     * <code>int32 total = 2;</code>
     * @return The total.
     */
    @java.lang.Override
    public int getTotal() {
      return total_;
    }
    /**
     * <code>int32 total = 2;</code>
     * @param value The total to set.
     * @return This builder for chaining.
     */
    public Builder setTotal(int value) {

      total_ = value;
      bitField0_ |= 0x00000002;
      onChanged();
      return this;
    }
    /**
     * <code>int32 total = 2;</code>
     * @return This builder for chaining.
     */
    public Builder clearTotal() {
      bitField0_ = (bitField0_ & ~0x00000002);
      total_ = 0;
      onChanged();
      return this;
    }

    // @@protoc_insertion_point(builder_scope:agent.ListAgentsGroupResponse)
  }

  // @@protoc_insertion_point(class_scope:agent.ListAgentsGroupResponse)
  private static final com.park.utmstack.service.grpc.ListAgentsGroupResponse DEFAULT_INSTANCE;
  static {
    DEFAULT_INSTANCE = new com.park.utmstack.service.grpc.ListAgentsGroupResponse();
  }

  public static com.park.utmstack.service.grpc.ListAgentsGroupResponse getDefaultInstance() {
    return DEFAULT_INSTANCE;
  }

  private static final com.google.protobuf.Parser<ListAgentsGroupResponse>
      PARSER = new com.google.protobuf.AbstractParser<ListAgentsGroupResponse>() {
    @java.lang.Override
    public ListAgentsGroupResponse parsePartialFrom(
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

  public static com.google.protobuf.Parser<ListAgentsGroupResponse> parser() {
    return PARSER;
  }

  @java.lang.Override
  public com.google.protobuf.Parser<ListAgentsGroupResponse> getParserForType() {
    return PARSER;
  }

  @java.lang.Override
  public com.park.utmstack.service.grpc.ListAgentsGroupResponse getDefaultInstanceForType() {
    return DEFAULT_INSTANCE;
  }

}

