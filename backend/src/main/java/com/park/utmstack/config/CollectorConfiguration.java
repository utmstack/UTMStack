package com.park.utmstack.config;

import com.utmstack.grpc.connection.GrpcConnection;
import com.utmstack.grpc.exception.GrpcConnectionException;
import com.utmstack.grpc.jclient.config.interceptors.impl.GrpcEmptyAuthInterceptor;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class CollectorConfiguration {
    private GrpcConnection collectorConnection;

    @Value("${grpc.server.address}")
    private String serverAddress;

    @Value("${grpc.server.port}")
    private Integer serverPort;

    @Bean
    public GrpcConnection collectorConnection() throws GrpcConnectionException {
        this.collectorConnection = new GrpcConnection();
        this.collectorConnection.createChannel(serverAddress, serverPort, new GrpcEmptyAuthInterceptor());
        return this.collectorConnection;
    }
}
