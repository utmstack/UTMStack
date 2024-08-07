package com.park.utmstack.config;

import com.park.utmstack.security.GrpcInterceptor;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.netty.GrpcSslContexts;
import io.grpc.netty.NettyChannelBuilder;
import io.netty.handler.ssl.util.InsecureTrustManagerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

import javax.annotation.PreDestroy;
import javax.net.ssl.SSLException;

@Configuration
public class GrpcConfiguration {
    private ManagedChannel channel;

    @Value("${grpc.server.address}")
    private String serverAddress;

    @Value("${grpc.server.port}")
    private Integer serverPort;

    @Bean
    public ManagedChannel managedChannel() throws SSLException {
        this.channel = NettyChannelBuilder
                .forAddress(serverAddress, serverPort)
                .intercept(new GrpcInterceptor())
                .sslContext(GrpcSslContexts.forClient().trustManager(InsecureTrustManagerFactory.INSTANCE).build())
                .build();
        return this.channel;
    }

    @PreDestroy
    public void shutdownChannel() {
        this.channel.shutdown();
    }
}
