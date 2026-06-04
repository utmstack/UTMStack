package com.park.utmstack.security;

import com.park.utmstack.config.Constants;
import io.grpc.*;

public class GrpcInterceptor implements ClientInterceptor {
    private static final Metadata.Key<String> CUSTOM_HEADER_KEY =
        Metadata.Key.of(Constants.AGENT_MANAGER_INTERNAL_KEY_HEADER, Metadata.ASCII_STRING_MARSHALLER);

    @Override
    public <ReqT, RespT> ClientCall<ReqT, RespT> interceptCall(MethodDescriptor<ReqT, RespT> method, CallOptions callOptions, Channel next) {
        return new ForwardingClientCall.SimpleForwardingClientCall<>(next.newCall(method, callOptions)) {
            @Override
            public void start(Listener<RespT> responseListener, Metadata headers) {
                headers.put(CUSTOM_HEADER_KEY, System.getenv(Constants.ENV_INTERNAL_KEY));
                super.start(responseListener, headers);
            }
        };
    }
}
