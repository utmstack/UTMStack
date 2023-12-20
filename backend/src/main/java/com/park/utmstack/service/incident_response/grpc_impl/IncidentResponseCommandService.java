package com.park.utmstack.service.incident_response.grpc_impl;

import com.park.utmstack.service.grpc.CommandResult;
import com.park.utmstack.service.grpc.PanelServiceGrpc;
import com.park.utmstack.service.grpc.UtmCommand;
import io.grpc.ManagedChannel;
import io.grpc.stub.StreamObserver;
import org.springframework.stereotype.Service;

@Service
public class IncidentResponseCommandService {

    private final PanelServiceGrpc.PanelServiceStub nonBlockingStub;

    public IncidentResponseCommandService(ManagedChannel grpcManagedChannel) {
        this.nonBlockingStub = PanelServiceGrpc.newStub(grpcManagedChannel);
    }

    public void sendCommand(String agentKey,
                            String command,
                            String originType,
                            String originId,
                            String reason,
                            String executedBy,
                            StreamObserver<CommandResult> responseObserver) {
        // Get the stub
        // Create the UtmCommand with the provided agent key and command

        UtmCommand utmCommand = UtmCommand.newBuilder()
            .setAgentKey(agentKey)
            .setCommand(command)
            .setOriginId(originId)
            .setOriginType(originType)
            .setReason(reason)
            .setExecutedBy(executedBy)
            .build();

        // Send command using the bidirectional stream
        StreamObserver<UtmCommand> requestObserver = nonBlockingStub.processCommand(responseObserver);
        try {
            requestObserver.onNext(utmCommand);
            // Mark the end of requests
            requestObserver.onCompleted();
        } catch (RuntimeException e) {
            // Cancel RPC
            requestObserver.onError(e);
            throw e;
        }
    }
}
