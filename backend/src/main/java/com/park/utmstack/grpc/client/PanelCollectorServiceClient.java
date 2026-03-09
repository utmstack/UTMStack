package com.park.utmstack.grpc.client;

import agent.CollectorOuterClass;
import agent.PanelCollectorServiceGrpc;
import io.grpc.ManagedChannel;
import io.grpc.StatusRuntimeException;

public class PanelCollectorServiceClient {

    private final ManagedChannel channel;
    private final PanelCollectorServiceGrpc.PanelCollectorServiceBlockingStub baseStub;

    public PanelCollectorServiceClient(ManagedChannel channel) {
        this.channel = channel;
        this.baseStub = PanelCollectorServiceGrpc.newBlockingStub(channel);
    }

    public CollectorOuterClass.ConfigKnowledge insertCollectorConfig(CollectorOuterClass.CollectorConfig config) {

        try {
            return baseStub.registerCollectorConfig(config);

        } catch (StatusRuntimeException e) {
            throw new RuntimeException("gRPC error inserting collector config: " + e.getMessage(), e);
        } catch (Exception e) {
            throw new RuntimeException("Unexpected error inserting collector config: " + e.getMessage(), e);
        }
    }

    public void shutdown() {
        channel.shutdown();
    }
}

