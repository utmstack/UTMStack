// Generated by the protocol buffer compiler.  DO NOT EDIT!
// NO CHECKED-IN PROTOBUF GENCODE
// source: common.proto
// Protobuf Java Version: 4.29.3

package com.park.utmstack.service.grpc;

/**
 * Protobuf enum {@code agent.ConnectorType}
 */
public enum ConnectorType
    implements com.google.protobuf.ProtocolMessageEnum {
  /**
   * <code>AGENT = 0;</code>
   */
  AGENT(0),
  /**
   * <code>COLLECTOR = 1;</code>
   */
  COLLECTOR(1),
  UNRECOGNIZED(-1),
  ;

  static {
    com.google.protobuf.RuntimeVersion.validateProtobufGencodeVersion(
      com.google.protobuf.RuntimeVersion.RuntimeDomain.PUBLIC,
      /* major= */ 4,
      /* minor= */ 29,
      /* patch= */ 3,
      /* suffix= */ "",
      ConnectorType.class.getName());
  }
  /**
   * <code>AGENT = 0;</code>
   */
  public static final int AGENT_VALUE = 0;
  /**
   * <code>COLLECTOR = 1;</code>
   */
  public static final int COLLECTOR_VALUE = 1;


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
  public static ConnectorType valueOf(int value) {
    return forNumber(value);
  }

  /**
   * @param value The numeric wire value of the corresponding enum entry.
   * @return The enum associated with the given numeric wire value.
   */
  public static ConnectorType forNumber(int value) {
    switch (value) {
      case 0: return AGENT;
      case 1: return COLLECTOR;
      default: return null;
    }
  }

  public static com.google.protobuf.Internal.EnumLiteMap<ConnectorType>
      internalGetValueMap() {
    return internalValueMap;
  }
  private static final com.google.protobuf.Internal.EnumLiteMap<
      ConnectorType> internalValueMap =
        new com.google.protobuf.Internal.EnumLiteMap<ConnectorType>() {
          public ConnectorType findValueByNumber(int number) {
            return ConnectorType.forNumber(number);
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
    return com.park.utmstack.service.grpc.Common.getDescriptor().getEnumTypes().get(1);
  }

  private static final ConnectorType[] VALUES = values();

  public static ConnectorType valueOf(
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

  private ConnectorType(int value) {
    this.value = value;
  }

  // @@protoc_insertion_point(enum_scope:agent.ConnectorType)
}

