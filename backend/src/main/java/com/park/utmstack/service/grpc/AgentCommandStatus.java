// Generated by the protocol buffer compiler.  DO NOT EDIT!
// NO CHECKED-IN PROTOBUF GENCODE
// source: agent.proto
// Protobuf Java Version: 4.29.3

package com.park.utmstack.service.grpc;

/**
 * Protobuf enum {@code agent.AgentCommandStatus}
 */
public enum AgentCommandStatus
    implements com.google.protobuf.ProtocolMessageEnum {
  /**
   * <code>NOT_EXECUTED = 0;</code>
   */
  NOT_EXECUTED(0),
  /**
   * <code>QUEUE = 1;</code>
   */
  QUEUE(1),
  /**
   * <code>PENDING = 2;</code>
   */
  PENDING(2),
  /**
   * <code>EXECUTED = 3;</code>
   */
  EXECUTED(3),
  /**
   * <code>ERROR = 4;</code>
   */
  ERROR(4),
  UNRECOGNIZED(-1),
  ;

  static {
    com.google.protobuf.RuntimeVersion.validateProtobufGencodeVersion(
      com.google.protobuf.RuntimeVersion.RuntimeDomain.PUBLIC,
      /* major= */ 4,
      /* minor= */ 29,
      /* patch= */ 3,
      /* suffix= */ "",
      AgentCommandStatus.class.getName());
  }
  /**
   * <code>NOT_EXECUTED = 0;</code>
   */
  public static final int NOT_EXECUTED_VALUE = 0;
  /**
   * <code>QUEUE = 1;</code>
   */
  public static final int QUEUE_VALUE = 1;
  /**
   * <code>PENDING = 2;</code>
   */
  public static final int PENDING_VALUE = 2;
  /**
   * <code>EXECUTED = 3;</code>
   */
  public static final int EXECUTED_VALUE = 3;
  /**
   * <code>ERROR = 4;</code>
   */
  public static final int ERROR_VALUE = 4;


  public final int getNumber() {
    if (this == UNRECOGNIZED) {
      throw new java.lang.IllegalArgumentException(
          "Can't get the number of an unknown enum value.");
    }
    return value;
  }

  /**
   * @param value The numeric wire value of the corresponding enum entry.
   * @return The enum associated with the given numeric wire value.
   * @deprecated Use {@link #forNumber(int)} instead.
   */
  @java.lang.Deprecated
  public static AgentCommandStatus valueOf(int value) {
    return forNumber(value);
  }

  /**
   * @param value The numeric wire value of the corresponding enum entry.
   * @return The enum associated with the given numeric wire value.
   */
  public static AgentCommandStatus forNumber(int value) {
    switch (value) {
      case 0: return NOT_EXECUTED;
      case 1: return QUEUE;
      case 2: return PENDING;
      case 3: return EXECUTED;
      case 4: return ERROR;
      default: return null;
    }
  }

  public static com.google.protobuf.Internal.EnumLiteMap<AgentCommandStatus>
      internalGetValueMap() {
    return internalValueMap;
  }
  private static final com.google.protobuf.Internal.EnumLiteMap<
      AgentCommandStatus> internalValueMap =
        new com.google.protobuf.Internal.EnumLiteMap<AgentCommandStatus>() {
          public AgentCommandStatus findValueByNumber(int number) {
            return AgentCommandStatus.forNumber(number);
          }
        };

  public final com.google.protobuf.Descriptors.EnumValueDescriptor
      getValueDescriptor() {
    if (this == UNRECOGNIZED) {
      throw new java.lang.IllegalStateException(
          "Can't get the descriptor of an unrecognized enum value.");
    }
    return getDescriptor().getValues().get(ordinal());
  }
  public final com.google.protobuf.Descriptors.EnumDescriptor
      getDescriptorForType() {
    return getDescriptor();
  }
  public static final com.google.protobuf.Descriptors.EnumDescriptor
      getDescriptor() {
    return com.park.utmstack.service.grpc.AgentManagerGrpc.getDescriptor().getEnumTypes().get(0);
  }

  private static final AgentCommandStatus[] VALUES = values();

  public static AgentCommandStatus valueOf(
      com.google.protobuf.Descriptors.EnumValueDescriptor desc) {
    if (desc.getType() != getDescriptor()) {
      throw new java.lang.IllegalArgumentException(
        "EnumValueDescriptor is not for this type.");
    }
    if (desc.getIndex() == -1) {
      return UNRECOGNIZED;
    }
    return VALUES[desc.getIndex()];
  }

  private final int value;

  private AgentCommandStatus(int value) {
    this.value = value;
  }

  // @@protoc_insertion_point(enum_scope:agent.AgentCommandStatus)
}

