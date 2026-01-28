package com.park.utmstack.grpc.interceptor;

import io.grpc.*;

public class CollectorAuthInterceptor implements ClientInterceptor {

    private static final Metadata.Key<String> ID_HEADER =
            Metadata.Key.of("id", Metadata.ASCII_STRING_MARSHALLER);

    private static final Metadata.Key<String> KEY_HEADER =
            Metadata.Key.of("key", Metadata.ASCII_STRING_MARSHALLER);

    private static final Metadata.Key<String> TYPE_HEADER =
            Metadata.Key.of("type", Metadata.ASCII_STRING_MARSHALLER);

    private final String collectorId;
    private final String collectorKey;

    public CollectorAuthInterceptor(String collectorId, String collectorKey) {
        this.collectorId = collectorId;
        this.collectorKey = collectorKey;
    }

    @Override
    public <ReqT, RespT> ClientCall<ReqT, RespT> interceptCall(
            MethodDescriptor<ReqT, RespT> methodDescriptor,
            CallOptions callOptions,
            Channel channel) {

        return new ForwardingClientCall.SimpleForwardingClientCall<>(
                channel.newCall(methodDescriptor, callOptions)) {

            @Override
            public void start(Listener<RespT> responseListener, Metadata headers) {

                headers.put(ID_HEADER, collectorId);
                headers.put(KEY_HEADER, collectorKey);
                headers.put(TYPE_HEADER, "collector");

                super.start(responseListener, headers);
            }
        };
    }
}

