// Generated by the protocol buffer compiler.  DO NOT EDIT!
// source: agent.proto

package com.park.utmstack.service.grpc;

public final class AgentManagerGrpc {
  private AgentManagerGrpc() {}
  public static void registerAllExtensions(
      com.google.protobuf.ExtensionRegistryLite registry) {
  }

  public static void registerAllExtensions(
      com.google.protobuf.ExtensionRegistry registry) {
    registerAllExtensions(
        (com.google.protobuf.ExtensionRegistryLite) registry);
  }
  static final com.google.protobuf.Descriptors.Descriptor
    internal_static_agent_Id_descriptor;
  static final 
    com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
      internal_static_agent_Id_fieldAccessorTable;
  static final com.google.protobuf.Descriptors.Descriptor
    internal_static_agent_Hostname_descriptor;
  static final 
    com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
      internal_static_agent_Hostname_fieldAccessorTable;
  static final com.google.protobuf.Descriptors.Descriptor
    internal_static_agent_Agent_descriptor;
  static final 
    com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
      internal_static_agent_Agent_fieldAccessorTable;
  static final com.google.protobuf.Descriptors.Descriptor
    internal_static_agent_AgentTypeUpdate_descriptor;
  static final 
    com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
      internal_static_agent_AgentTypeUpdate_fieldAccessorTable;
  static final com.google.protobuf.Descriptors.Descriptor
    internal_static_agent_AgentGroupUpdate_descriptor;
  static final 
    com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
      internal_static_agent_AgentGroupUpdate_fieldAccessorTable;
  static final com.google.protobuf.Descriptors.Descriptor
    internal_static_agent_AgentRequest_descriptor;
  static final 
    com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
      internal_static_agent_AgentRequest_fieldAccessorTable;
  static final com.google.protobuf.Descriptors.Descriptor
    internal_static_agent_AgentResponse_descriptor;
  static final 
    com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
      internal_static_agent_AgentResponse_fieldAccessorTable;
  static final com.google.protobuf.Descriptors.Descriptor
    internal_static_agent_AgentDelete_descriptor;
  static final 
    com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
      internal_static_agent_AgentDelete_fieldAccessorTable;
  static final com.google.protobuf.Descriptors.Descriptor
    internal_static_agent_UtmCommand_descriptor;
  static final 
    com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
      internal_static_agent_UtmCommand_fieldAccessorTable;
  static final com.google.protobuf.Descriptors.Descriptor
    internal_static_agent_CommandResult_descriptor;
  static final 
    com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
      internal_static_agent_CommandResult_fieldAccessorTable;
  static final com.google.protobuf.Descriptors.Descriptor
    internal_static_agent_PingRequest_descriptor;
  static final 
    com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
      internal_static_agent_PingRequest_fieldAccessorTable;
  static final com.google.protobuf.Descriptors.Descriptor
    internal_static_agent_PingResponse_descriptor;
  static final 
    com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
      internal_static_agent_PingResponse_fieldAccessorTable;
  static final com.google.protobuf.Descriptors.Descriptor
    internal_static_agent_BidirectionalStream_descriptor;
  static final 
    com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
      internal_static_agent_BidirectionalStream_fieldAccessorTable;
  static final com.google.protobuf.Descriptors.Descriptor
    internal_static_agent_AuthResponse_descriptor;
  static final 
    com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
      internal_static_agent_AuthResponse_fieldAccessorTable;
  static final com.google.protobuf.Descriptors.Descriptor
    internal_static_agent_ListRequest_descriptor;
  static final 
    com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
      internal_static_agent_ListRequest_fieldAccessorTable;
  static final com.google.protobuf.Descriptors.Descriptor
    internal_static_agent_ListAgentsResponse_descriptor;
  static final 
    com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
      internal_static_agent_ListAgentsResponse_fieldAccessorTable;
  static final com.google.protobuf.Descriptors.Descriptor
    internal_static_agent_ListAgentsCommandsResponse_descriptor;
  static final 
    com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
      internal_static_agent_ListAgentsCommandsResponse_fieldAccessorTable;
  static final com.google.protobuf.Descriptors.Descriptor
    internal_static_agent_AgentCommand_descriptor;
  static final 
    com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
      internal_static_agent_AgentCommand_fieldAccessorTable;
  static final com.google.protobuf.Descriptors.Descriptor
    internal_static_agent_AgentType_descriptor;
  static final 
    com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
      internal_static_agent_AgentType_fieldAccessorTable;
  static final com.google.protobuf.Descriptors.Descriptor
    internal_static_agent_AgentGroup_descriptor;
  static final 
    com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
      internal_static_agent_AgentGroup_fieldAccessorTable;
  static final com.google.protobuf.Descriptors.Descriptor
    internal_static_agent_ListAgentsGroupResponse_descriptor;
  static final 
    com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
      internal_static_agent_ListAgentsGroupResponse_fieldAccessorTable;
  static final com.google.protobuf.Descriptors.Descriptor
    internal_static_agent_ListAgentsTypeResponse_descriptor;
  static final 
    com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
      internal_static_agent_ListAgentsTypeResponse_fieldAccessorTable;

  public static com.google.protobuf.Descriptors.FileDescriptor
      getDescriptor() {
    return descriptor;
  }
  private static  com.google.protobuf.Descriptors.FileDescriptor
      descriptor;
  static {
    java.lang.String[] descriptorData = {
      "\n\013agent.proto\022\005agent\032\037google/protobuf/ti" +
      "mestamp.proto\"\020\n\002Id\022\n\n\002id\030\001 \001(\003\"\034\n\010Hostn" +
      "ame\022\020\n\010hostname\030\001 \001(\t\"\217\002\n\005Agent\022\n\n\002ip\030\001 " +
      "\001(\t\022\020\n\010hostname\030\002 \001(\t\022\n\n\002os\030\003 \001(\t\022\"\n\006sta" +
      "tus\030\004 \001(\0162\022.agent.AgentStatus\022\020\n\010platfor" +
      "m\030\005 \001(\t\022\017\n\007version\030\006 \001(\t\022\021\n\tagent_key\030\007 " +
      "\001(\t\022\n\n\002id\030\010 \001(\r\022\021\n\tlast_seen\030\t \001(\t\022\013\n\003ma" +
      "c\030\n \001(\t\022\030\n\020os_major_version\030\013 \001(\t\022\030\n\020os_" +
      "minor_version\030\014 \001(\t\022\017\n\007aliases\030\r \001(\t\022\021\n\t" +
      "addresses\030\016 \001(\t\"7\n\017AgentTypeUpdate\022\020\n\010ag" +
      "ent_id\030\001 \001(\r\022\022\n\nagent_type\030\002 \001(\r\"9\n\020Agen" +
      "tGroupUpdate\022\020\n\010agent_id\030\001 \001(\r\022\023\n\013agent_" +
      "group\030\003 \001(\r\"\325\001\n\014AgentRequest\022\n\n\002ip\030\001 \001(\t" +
      "\022\020\n\010hostname\030\002 \001(\t\022\n\n\002os\030\003 \001(\t\022\020\n\010platfo" +
      "rm\030\004 \001(\t\022\017\n\007version\030\005 \001(\t\022\023\n\013register_by" +
      "\030\006 \001(\t\022\013\n\003mac\030\007 \001(\t\022\030\n\020os_major_version\030" +
      "\010 \001(\t\022\030\n\020os_minor_version\030\t \001(\t\022\017\n\007alias" +
      "es\030\n \001(\t\022\021\n\taddresses\030\013 \001(\t\".\n\rAgentResp" +
      "onse\022\n\n\002id\030\001 \001(\r\022\021\n\tagent_key\030\002 \001(\t\"4\n\013A" +
      "gentDelete\022\022\n\ndeleted_by\030\001 \001(\t\022\021\n\tagent_" +
      "key\030\002 \001(\t\"\215\001\n\nUtmCommand\022\021\n\tagent_key\030\001 " +
      "\001(\t\022\017\n\007command\030\002 \001(\t\022\023\n\013executed_by\030\003 \001(" +
      "\t\022\016\n\006cmd_id\030\004 \001(\t\022\023\n\013origin_type\030\005 \001(\t\022\021" +
      "\n\torigin_id\030\006 \001(\t\022\016\n\006reason\030\007 \001(\t\"s\n\rCom" +
      "mandResult\022\021\n\tagent_key\030\001 \001(\t\022\016\n\006result\030" +
      "\002 \001(\t\022/\n\013executed_at\030\003 \001(\0132\032.google.prot" +
      "obuf.Timestamp\022\016\n\006cmd_id\030\004 \001(\t\" \n\013PingRe" +
      "quest\022\021\n\tagent_key\030\001 \001(\t\"3\n\014PingResponse" +
      "\022\020\n\010is_alive\030\001 \001(\010\022\021\n\tagent_key\030\002 \001(\t\"\243\001" +
      "\n\023BidirectionalStream\022$\n\007command\030\001 \001(\0132\021" +
      ".agent.UtmCommandH\000\022&\n\006result\030\002 \001(\0132\024.ag" +
      "ent.CommandResultH\000\022,\n\rauth_response\030\003 \001" +
      "(\0132\023.agent.AuthResponseH\000B\020\n\016stream_mess" +
      "age\"3\n\014AuthResponse\022\020\n\010agent_id\030\001 \001(\004\022\021\n" +
      "\tagent_key\030\002 \001(\t\"\\\n\013ListRequest\022\023\n\013page_" +
      "number\030\001 \001(\005\022\021\n\tpage_size\030\002 \001(\005\022\024\n\014searc" +
      "h_query\030\003 \001(\t\022\017\n\007sort_by\030\004 \001(\t\"?\n\022ListAg" +
      "entsResponse\022\032\n\004rows\030\001 \003(\0132\014.agent.Agent" +
      "\022\r\n\005total\030\002 \001(\005\"N\n\032ListAgentsCommandsRes" +
      "ponse\022!\n\004rows\030\001 \003(\0132\023.agent.AgentCommand" +
      "\022\r\n\005total\030\002 \001(\005\"\316\002\n\014AgentCommand\022.\n\ncrea" +
      "ted_at\030\001 \001(\0132\032.google.protobuf.Timestamp" +
      "\022.\n\nupdated_at\030\002 \001(\0132\032.google.protobuf.T" +
      "imestamp\022\020\n\010agent_id\030\003 \001(\r\022\017\n\007command\030\004 " +
      "\001(\t\0221\n\016command_status\030\005 \001(\0162\031.agent.Agen" +
      "tCommandStatus\022\016\n\006result\030\006 \001(\t\022\023\n\013execut" +
      "ed_by\030\007 \001(\t\022\016\n\006cmd_id\030\010 \001(\t\022\033\n\005agent\030\t \001" +
      "(\0132\014.agent.Agent\022\016\n\006reason\030\n \001(\t\022\023\n\013orig" +
      "in_type\030\013 \001(\t\022\021\n\torigin_id\030\014 \001(\t\"*\n\tAgen" +
      "tType\022\n\n\002id\030\001 \001(\r\022\021\n\ttype_name\030\002 \001(\t\"G\n\n" +
      "AgentGroup\022\n\n\002id\030\001 \001(\r\022\022\n\ngroup_name\030\002 \001" +
      "(\t\022\031\n\021group_description\030\003 \001(\t\"I\n\027ListAge" +
      "ntsGroupResponse\022\037\n\004rows\030\001 \003(\0132\021.agent.A" +
      "gentGroup\022\r\n\005total\030\002 \001(\005\"G\n\026ListAgentsTy" +
      "peResponse\022\036\n\004rows\030\001 \003(\0132\020.agent.AgentTy" +
      "pe\022\r\n\005total\030\002 \001(\005*3\n\013AgentStatus\022\n\n\006ONLI" +
      "NE\020\000\022\013\n\007OFFLINE\020\001\022\013\n\007UNKNOWN\020\002*W\n\022AgentC" +
      "ommandStatus\022\020\n\014NOT_EXECUTED\020\000\022\t\n\005QUEUE\020" +
      "\001\022\013\n\007PENDING\020\002\022\014\n\010EXECUTED\020\003\022\t\n\005ERROR\020\0042" +
      "\235\004\n\014AgentService\022K\n\013AgentStream\022\032.agent." +
      "BidirectionalStream\032\032.agent.Bidirectiona" +
      "lStream\"\000(\0010\001\022=\n\nListAgents\022\022.agent.List" +
      "Request\032\031.agent.ListAgentsResponse\"\000\0229\n\017" +
      "UpdateAgentType\022\026.agent.AgentTypeUpdate\032" +
      "\014.agent.Agent\"\000\022;\n\020UpdateAgentGroup\022\027.ag" +
      "ent.AgentGroupUpdate\032\014.agent.Agent\"\000\022L\n\021" +
      "ListAgentCommands\022\022.agent.ListRequest\032!." +
      "agent.ListAgentsCommandsResponse\"\000\0225\n\022Ge" +
      "tAgentByHostname\022\017.agent.Hostname\032\014.agen" +
      "t.Agent\"\000\022I\n\026ListAgentsWithCommands\022\022.ag" +
      "ent.ListRequest\032\031.agent.ListAgentsRespon" +
      "se\"\000\0229\n\013DeleteAgent\022\022.agent.AgentDelete\032" +
      "\024.agent.AgentResponse\"\0002O\n\014PanelService\022" +
      "?\n\016ProcessCommand\022\021.agent.UtmCommand\032\024.a" +
      "gent.CommandResult\"\000(\0010\0012\352\001\n\021AgentGroupS" +
      "ervice\0225\n\013CreateGroup\022\021.agent.AgentGroup" +
      "\032\021.agent.AgentGroup\"\000\0223\n\tEditGroup\022\021.age" +
      "nt.AgentGroup\032\021.agent.AgentGroup\"\000\022B\n\nLi" +
      "stGroups\022\022.agent.ListRequest\032\036.agent.Lis" +
      "tAgentsGroupResponse\"\000\022%\n\013DeleteGroup\022\t." +
      "agent.Id\032\t.agent.Id\"\0002Y\n\020AgentTypeServic" +
      "e\022E\n\016ListAgentTypes\022\022.agent.ListRequest\032" +
      "\035.agent.ListAgentsTypeResponse\"\000B7\n\036com." +
      "park.utmstack.service.grpcB\020AgentManager" +
      "GrpcP\001\210\001\001b\006proto3"
    };
    descriptor = com.google.protobuf.Descriptors.FileDescriptor
      .internalBuildGeneratedFileFrom(descriptorData,
        new com.google.protobuf.Descriptors.FileDescriptor[] {
          com.google.protobuf.TimestampProto.getDescriptor(),
        });
    internal_static_agent_Id_descriptor =
      getDescriptor().getMessageTypes().get(0);
    internal_static_agent_Id_fieldAccessorTable = new
      com.google.protobuf.GeneratedMessageV3.FieldAccessorTable(
        internal_static_agent_Id_descriptor,
        new java.lang.String[] { "Id", });
    internal_static_agent_Hostname_descriptor =
      getDescriptor().getMessageTypes().get(1);
    internal_static_agent_Hostname_fieldAccessorTable = new
      com.google.protobuf.GeneratedMessageV3.FieldAccessorTable(
        internal_static_agent_Hostname_descriptor,
        new java.lang.String[] { "Hostname", });
    internal_static_agent_Agent_descriptor =
      getDescriptor().getMessageTypes().get(2);
    internal_static_agent_Agent_fieldAccessorTable = new
      com.google.protobuf.GeneratedMessageV3.FieldAccessorTable(
        internal_static_agent_Agent_descriptor,
        new java.lang.String[] { "Ip", "Hostname", "Os", "Status", "Platform", "Version", "AgentKey", "Id", "LastSeen", "Mac", "OsMajorVersion", "OsMinorVersion", "Aliases", "Addresses", });
    internal_static_agent_AgentTypeUpdate_descriptor =
      getDescriptor().getMessageTypes().get(3);
    internal_static_agent_AgentTypeUpdate_fieldAccessorTable = new
      com.google.protobuf.GeneratedMessageV3.FieldAccessorTable(
        internal_static_agent_AgentTypeUpdate_descriptor,
        new java.lang.String[] { "AgentId", "AgentType", });
    internal_static_agent_AgentGroupUpdate_descriptor =
      getDescriptor().getMessageTypes().get(4);
    internal_static_agent_AgentGroupUpdate_fieldAccessorTable = new
      com.google.protobuf.GeneratedMessageV3.FieldAccessorTable(
        internal_static_agent_AgentGroupUpdate_descriptor,
        new java.lang.String[] { "AgentId", "AgentGroup", });
    internal_static_agent_AgentRequest_descriptor =
      getDescriptor().getMessageTypes().get(5);
    internal_static_agent_AgentRequest_fieldAccessorTable = new
      com.google.protobuf.GeneratedMessageV3.FieldAccessorTable(
        internal_static_agent_AgentRequest_descriptor,
        new java.lang.String[] { "Ip", "Hostname", "Os", "Platform", "Version", "RegisterBy", "Mac", "OsMajorVersion", "OsMinorVersion", "Aliases", "Addresses", });
    internal_static_agent_AgentResponse_descriptor =
      getDescriptor().getMessageTypes().get(6);
    internal_static_agent_AgentResponse_fieldAccessorTable = new
      com.google.protobuf.GeneratedMessageV3.FieldAccessorTable(
        internal_static_agent_AgentResponse_descriptor,
        new java.lang.String[] { "Id", "AgentKey", });
    internal_static_agent_AgentDelete_descriptor =
      getDescriptor().getMessageTypes().get(7);
    internal_static_agent_AgentDelete_fieldAccessorTable = new
      com.google.protobuf.GeneratedMessageV3.FieldAccessorTable(
        internal_static_agent_AgentDelete_descriptor,
        new java.lang.String[] { "DeletedBy", "AgentKey", });
    internal_static_agent_UtmCommand_descriptor =
      getDescriptor().getMessageTypes().get(8);
    internal_static_agent_UtmCommand_fieldAccessorTable = new
      com.google.protobuf.GeneratedMessageV3.FieldAccessorTable(
        internal_static_agent_UtmCommand_descriptor,
        new java.lang.String[] { "AgentKey", "Command", "ExecutedBy", "CmdId", "OriginType", "OriginId", "Reason", });
    internal_static_agent_CommandResult_descriptor =
      getDescriptor().getMessageTypes().get(9);
    internal_static_agent_CommandResult_fieldAccessorTable = new
      com.google.protobuf.GeneratedMessageV3.FieldAccessorTable(
        internal_static_agent_CommandResult_descriptor,
        new java.lang.String[] { "AgentKey", "Result", "ExecutedAt", "CmdId", });
    internal_static_agent_PingRequest_descriptor =
      getDescriptor().getMessageTypes().get(10);
    internal_static_agent_PingRequest_fieldAccessorTable = new
      com.google.protobuf.GeneratedMessageV3.FieldAccessorTable(
        internal_static_agent_PingRequest_descriptor,
        new java.lang.String[] { "AgentKey", });
    internal_static_agent_PingResponse_descriptor =
      getDescriptor().getMessageTypes().get(11);
    internal_static_agent_PingResponse_fieldAccessorTable = new
      com.google.protobuf.GeneratedMessageV3.FieldAccessorTable(
        internal_static_agent_PingResponse_descriptor,
        new java.lang.String[] { "IsAlive", "AgentKey", });
    internal_static_agent_BidirectionalStream_descriptor =
      getDescriptor().getMessageTypes().get(12);
    internal_static_agent_BidirectionalStream_fieldAccessorTable = new
      com.google.protobuf.GeneratedMessageV3.FieldAccessorTable(
        internal_static_agent_BidirectionalStream_descriptor,
        new java.lang.String[] { "Command", "Result", "AuthResponse", "StreamMessage", });
    internal_static_agent_AuthResponse_descriptor =
      getDescriptor().getMessageTypes().get(13);
    internal_static_agent_AuthResponse_fieldAccessorTable = new
      com.google.protobuf.GeneratedMessageV3.FieldAccessorTable(
        internal_static_agent_AuthResponse_descriptor,
        new java.lang.String[] { "AgentId", "AgentKey", });
    internal_static_agent_ListRequest_descriptor =
      getDescriptor().getMessageTypes().get(14);
    internal_static_agent_ListRequest_fieldAccessorTable = new
      com.google.protobuf.GeneratedMessageV3.FieldAccessorTable(
        internal_static_agent_ListRequest_descriptor,
        new java.lang.String[] { "PageNumber", "PageSize", "SearchQuery", "SortBy", });
    internal_static_agent_ListAgentsResponse_descriptor =
      getDescriptor().getMessageTypes().get(15);
    internal_static_agent_ListAgentsResponse_fieldAccessorTable = new
      com.google.protobuf.GeneratedMessageV3.FieldAccessorTable(
        internal_static_agent_ListAgentsResponse_descriptor,
        new java.lang.String[] { "Rows", "Total", });
    internal_static_agent_ListAgentsCommandsResponse_descriptor =
      getDescriptor().getMessageTypes().get(16);
    internal_static_agent_ListAgentsCommandsResponse_fieldAccessorTable = new
      com.google.protobuf.GeneratedMessageV3.FieldAccessorTable(
        internal_static_agent_ListAgentsCommandsResponse_descriptor,
        new java.lang.String[] { "Rows", "Total", });
    internal_static_agent_AgentCommand_descriptor =
      getDescriptor().getMessageTypes().get(17);
    internal_static_agent_AgentCommand_fieldAccessorTable = new
      com.google.protobuf.GeneratedMessageV3.FieldAccessorTable(
        internal_static_agent_AgentCommand_descriptor,
        new java.lang.String[] { "CreatedAt", "UpdatedAt", "AgentId", "Command", "CommandStatus", "Result", "ExecutedBy", "CmdId", "Agent", "Reason", "OriginType", "OriginId", });
    internal_static_agent_AgentType_descriptor =
      getDescriptor().getMessageTypes().get(18);
    internal_static_agent_AgentType_fieldAccessorTable = new
      com.google.protobuf.GeneratedMessageV3.FieldAccessorTable(
        internal_static_agent_AgentType_descriptor,
        new java.lang.String[] { "Id", "TypeName", });
    internal_static_agent_AgentGroup_descriptor =
      getDescriptor().getMessageTypes().get(19);
    internal_static_agent_AgentGroup_fieldAccessorTable = new
      com.google.protobuf.GeneratedMessageV3.FieldAccessorTable(
        internal_static_agent_AgentGroup_descriptor,
        new java.lang.String[] { "Id", "GroupName", "GroupDescription", });
    internal_static_agent_ListAgentsGroupResponse_descriptor =
      getDescriptor().getMessageTypes().get(20);
    internal_static_agent_ListAgentsGroupResponse_fieldAccessorTable = new
      com.google.protobuf.GeneratedMessageV3.FieldAccessorTable(
        internal_static_agent_ListAgentsGroupResponse_descriptor,
        new java.lang.String[] { "Rows", "Total", });
    internal_static_agent_ListAgentsTypeResponse_descriptor =
      getDescriptor().getMessageTypes().get(21);
    internal_static_agent_ListAgentsTypeResponse_fieldAccessorTable = new
      com.google.protobuf.GeneratedMessageV3.FieldAccessorTable(
        internal_static_agent_ListAgentsTypeResponse_descriptor,
        new java.lang.String[] { "Rows", "Total", });
    com.google.protobuf.TimestampProto.getDescriptor();
  }

  // @@protoc_insertion_point(outer_class_scope)
}
