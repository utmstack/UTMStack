package com.park.utmstack.grpc.connection;

import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import lombok.RequiredArgsConstructor;

@RequiredArgsConstructor
public class GrpcConnection {

    private ManagedChannel channel;
    private final String host;
    private final int port;

    public void connect() {
        this.channel = ManagedChannelBuilder
                .forAddress(host, port)
                .usePlaintext()
                .build();
    }

    public ManagedChannel getChannel() {
        if (channel == null) {
            throw new IllegalStateException("Channel not initialized. Call connect() first.");
        }
        return channel;
    }

    public void shutdown() {
        if (channel != null && !channel.isShutdown()) {
            channel.shutdown();
        }
    }
}

