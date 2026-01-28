package com.park.utmstack.config;

import com.park.utmstack.grpc.connection.GrpcConnection;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class CollectorConfiguration {

    @Value("${grpc.server.address}")
    private String serverAddress;

    @Value("${grpc.server.port}")
    private Integer serverPort;

    @Bean
    public GrpcConnection collectorConnection()  {

        GrpcConnection collectorConnection = new GrpcConnection(this.serverAddress, this.serverPort);
        collectorConnection.connect();

        return collectorConnection;
    }
}
