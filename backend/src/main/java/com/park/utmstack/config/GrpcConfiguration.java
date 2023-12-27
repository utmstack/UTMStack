package com.park.utmstack.config;

import com.park.utmstack.security.GrpcInterceptor;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

import javax.annotation.PreDestroy;

@Configuration
public class GrpcConfiguration {
    private ManagedChannel channel;

    @Value("${grpc.server.address}")
    private String serverAddress;

    @Value("${grpc.server.port}")
    private Integer serverPort;

    @Bean
    public ManagedChannel managedChannel() {
        this.channel = ManagedChannelBuilder.forAddress(serverAddress, serverPort)
            .intercept(new GrpcInterceptor())
            .usePlaintext()
            .enableRetry()
            .build();
        return this.channel;
    }

    @PreDestroy
    public void shutdownChannel() {
        this.channel.shutdown();
    }
}
