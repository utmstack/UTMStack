package com.park.utmstack.service.app_info;

import com.sun.net.httpserver.HttpServer;
import org.junit.jupiter.api.Test;

import java.net.InetSocketAddress;
import java.util.concurrent.atomic.AtomicInteger;

import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

class AppInfoServiceTest {

    @Test
    void returnsTrueAndCachesWhenCustomerManagerHealthResponds200() throws Exception {
        AtomicInteger hits = new AtomicInteger(0);
        HttpServer server = HttpServer.create(new InetSocketAddress("127.0.0.1", 0), 0);
        server.createContext("/api/v1/health", exchange -> {
            hits.incrementAndGet();
            exchange.sendResponseHeaders(200, -1);
            exchange.close();
        });
        server.start();
        try {
            String url = "http://127.0.0.1:" + server.getAddress().getPort() + "/api/v1/health";
            AppInfoService svc = new AppInfoService(url);

            assertTrue(svc.isCustomerManagerReachable());
            assertTrue(svc.isCustomerManagerReachable());
            assertTrue(hits.get() == 1, "Expected 60s cache to avoid duplicate probes, hits=" + hits.get());
        } finally {
            server.stop(0);
        }
    }

    @Test
    void returnsFalseWhenCustomerManagerHealthFailsOrUnreachable() throws Exception {
        HttpServer server = HttpServer.create(new InetSocketAddress("127.0.0.1", 0), 0);
        server.createContext("/api/v1/health", exchange -> {
            exchange.sendResponseHeaders(503, -1);
            exchange.close();
        });
        server.start();
        int port = server.getAddress().getPort();
        try {
            AppInfoService svc503 = new AppInfoService("http://127.0.0.1:" + port + "/api/v1/health");
            assertFalse(svc503.isCustomerManagerReachable());
        } finally {
            server.stop(0);
        }

        AppInfoService svcClosed = new AppInfoService("http://127.0.0.1:" + port + "/api/v1/health");
        assertFalse(svcClosed.isCustomerManagerReachable());
    }
}
