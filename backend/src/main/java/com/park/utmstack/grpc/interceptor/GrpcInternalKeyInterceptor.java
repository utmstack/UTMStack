package com.park.utmstack.grpc.interceptor;

import com.park.utmstack.config.Constants;
import io.grpc.*;

public class GrpcInternalKeyInterceptor implements ClientInterceptor {

    private static final Metadata.Key<String> INTERNAL_KEY_HEADER = Metadata.Key.of("internal-key", Metadata.ASCII_STRING_MARSHALLER);
    private static final Metadata.Key<String> TYPE_HEADER = Metadata.Key.of("type", Metadata.ASCII_STRING_MARSHALLER);

    @Override
    public <ReqT, RespT> ClientCall<ReqT, RespT> interceptCall(MethodDescriptor<ReqT, RespT> methodDescriptor, CallOptions callOptions, Channel channel) {

        return new ForwardingClientCall.SimpleForwardingClientCall<>(
                channel.newCall(methodDescriptor, callOptions)) {

            @Override public void start(Listener<RespT> responseListener, Metadata headers) {
                String internalKey = System.getenv(Constants.ENV_INTERNAL_KEY);

                headers.put(INTERNAL_KEY_HEADER, internalKey);
                headers.put(TYPE_HEADER, "internal");
                super.start(responseListener, headers);
            } };
    }
}
