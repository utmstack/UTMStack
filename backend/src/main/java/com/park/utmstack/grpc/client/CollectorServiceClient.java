package com.park.utmstack.grpc.client;

import agent.CollectorOuterClass;
import agent.CollectorServiceGrpc;
import com.park.utmstack.grpc.interceptor.CollectorAuthInterceptor;
import com.park.utmstack.grpc.interceptor.GrpcInternalKeyInterceptor;
import com.park.utmstack.service.grpc.AuthResponse;
import com.park.utmstack.service.grpc.DeleteRequest;
import com.park.utmstack.service.grpc.ListRequest;
import com.park.utmstack.util.exceptions.ApiException;
import io.grpc.ManagedChannel;
import io.grpc.StatusRuntimeException;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpStatus;

@Slf4j
public class CollectorServiceClient {

    private final ManagedChannel channel;
    private final CollectorServiceGrpc.CollectorServiceBlockingStub baseStub;

    public CollectorServiceClient(ManagedChannel channel) {
        this.channel = channel;
        this.baseStub = CollectorServiceGrpc.newBlockingStub(channel);
    }

    public CollectorOuterClass.ListCollectorResponse listCollectors(ListRequest request) {
        String ctx = "CollectorServiceClient.listCollectors";
        try {

            CollectorServiceGrpc.CollectorServiceBlockingStub stub =
                    baseStub.withInterceptors(new GrpcInternalKeyInterceptor());

            return stub.listCollector(request);

        } catch (StatusRuntimeException e) {
            log.error("{}: An error occurred while listing collectors: {}", ctx, e.getMessage());
            throw new ApiException(String.format("%s: gRPC error listing collectors", ctx), HttpStatus.INTERNAL_SERVER_ERROR);
        }
    }

    public AuthResponse deleteCollector(int collectorId, String collectorKey) {

        try {
            CollectorServiceGrpc.CollectorServiceBlockingStub stub =
                    baseStub.withInterceptors(
                            new CollectorAuthInterceptor(
                                    String.valueOf(collectorId),
                                    collectorKey
                            )
                    );

            DeleteRequest request = DeleteRequest.newBuilder()
                            .setDeletedBy(String.valueOf(collectorId))
                            .build();

            return stub.deleteCollector(request);

        } catch (StatusRuntimeException e) {
            log.error("{}: An error occurred while deleting collector:{}", collectorId, e.getMessage());
            throw new ApiException(String.format("%s: gRPC error deleting collector", collectorId), HttpStatus.INTERNAL_SERVER_ERROR);
        }
    }


    public CollectorOuterClass.CollectorConfig getCollectorConfig(int collectorId, String collectorKey, CollectorOuterClass.CollectorModule module) {

        try {
            CollectorServiceGrpc.CollectorServiceBlockingStub stub =
                    baseStub.withInterceptors(
                            new CollectorAuthInterceptor(
                                    String.valueOf(collectorId),
                                    collectorKey
                            )
                    );

            CollectorOuterClass.ConfigRequest request =
                    CollectorOuterClass.ConfigRequest.newBuilder()
                            .setModule(module)
                            .build();

            return stub.getCollectorConfig(request);

        } catch (StatusRuntimeException e) {
            log.error("{}: An error occurred while getting collector:{}", collectorId, e.getMessage());
            throw new ApiException(String.format("%s: gRPC error getting collector config", collectorId), HttpStatus.INTERNAL_SERVER_ERROR);
        }
    }

    public void shutdown() {
        channel.shutdown();
    }
}

