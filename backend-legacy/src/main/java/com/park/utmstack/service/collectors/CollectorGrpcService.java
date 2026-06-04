package com.park.utmstack.service.collectors;

import agent.CollectorOuterClass.*;
import com.park.utmstack.grpc.client.CollectorServiceClient;
import com.park.utmstack.grpc.client.PanelCollectorServiceClient;
import com.park.utmstack.service.grpc.ListRequest;
import io.grpc.ManagedChannel;
import org.springframework.stereotype.Service;

@Service
public class CollectorGrpcService {

    private final CollectorServiceClient collectorClient;
    private final PanelCollectorServiceClient panelClient;

    public CollectorGrpcService(ManagedChannel channel) {
        this.collectorClient = new CollectorServiceClient(channel);
        this.panelClient = new PanelCollectorServiceClient(channel);
    }

    public ListCollectorResponse listCollectors(ListRequest request) {
        return collectorClient.listCollectors(request);
    }

    public CollectorConfig getCollectorConfig(int id, String key, CollectorModule module) {
        return collectorClient.getCollectorConfig(id, key, module);
    }

    public void deleteCollector(int id, String key) {
        collectorClient.deleteCollector(id, key);
    }

    public ConfigKnowledge upsertCollectorConfig(CollectorConfig config) {
        return panelClient.insertCollectorConfig(config);
    }
}
