package com.park.utmstack.service.elasticsearch;

import com.park.utmstack.config.Constants;
import com.utmstack.opensearch_connector.OpenSearch;
import com.utmstack.opensearch_connector.enums.HttpScheme;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.context.event.ApplicationReadyEvent;
import org.springframework.context.event.EventListener;
import org.springframework.core.Ordered;
import org.springframework.core.annotation.Order;
import org.springframework.stereotype.Service;
import org.springframework.util.Assert;

@Service
public class OpensearchClientBuilder {
    private static final String CLASSNAME = "OpensearchClientBuilder";
    private final Logger log = LoggerFactory.getLogger(OpensearchClientBuilder.class);
    private volatile OpenSearch client;

    @FunctionalInterface
    public interface OsAction<T> {
        T apply(OpenSearch client) throws Exception;
    }

    @Order(Ordered.HIGHEST_PRECEDENCE)
    @EventListener(ApplicationReadyEvent.class)
    public void init() throws Exception {
        buildClient();
    }

    public OpenSearch getClient() {
        return client;
    }

    /**
     * Runs an action against the OpenSearch client with one-shot recovery: if the underlying
     * Apache HttpAsyncClient I/O reactor has transitioned to STOPPED (typically after an OOM
     * or a fatal callback exception while streaming a very large response), the singleton
     * client is rebuilt and the action is retried once. All other failures propagate unchanged.
     * Callers that don't need recovery should keep using {@link #getClient()} directly.
     */
    public <T> T execute(OsAction<T> action) throws Exception {
        try {
            return action.apply(client);
        } catch (Exception e) {
            if (!isReactorStopped(e))
                throw e;
            log.warn("OpenSearch I/O reactor is STOPPED; rebuilding client and retrying once", e);
            rebuild();
            return action.apply(client);
        }
    }

    public synchronized void rebuild() {
        final String ctx = CLASSNAME + ".rebuild";
        try {
            OpenSearch old = this.client;
            buildClient();
            tryClose(old);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            throw new RuntimeException(msg);
        }
    }

    private synchronized void buildClient() {
        final String ctx = CLASSNAME + ".buildClient";
        try {
            String host = System.getenv(Constants.ENV_ELASTICSEARCH_HOST);
            Assert.hasText(host, "Environment variable ELASTICSEARCH_HOST is missing or its value is null or empty");

            String port = System.getenv(Constants.ENV_ELASTICSEARCH_PORT);
            Assert.hasText(port, "Environment variable ELASTICSEARCH_PORT is missing or its value is null or empty");

            String user = System.getenv(Constants.ENV_ELASTICSEARCH_USER);
            Assert.hasText(user, "Environment variable ELASTICSEARCH_USER is missing or its value is null or empty");

            String password = System.getenv(Constants.ENV_ELASTICSEARCH_PASSWORD);
            Assert.hasText(password, "Environment variable ELASTICSEARCH_PASSWORD is missing or its value is null or empty");

            this.client = OpenSearch.builder()
                    .withHost(host, Integer.parseInt(port), HttpScheme.https)
                    .withCredentials(user, password)
                    .build();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            throw new RuntimeException(msg);
        }
    }

    private void tryClose(OpenSearch old) {
        if (old == null) return;
        try {
            if (old instanceof AutoCloseable) {
                ((AutoCloseable) old).close();
            }
        } catch (Exception ignored) {
            // best-effort: the old client is unusable anyway
        }
    }

    /**
     * Detects the Apache HttpAsyncClient "Request cannot be executed; I/O reactor status: STOPPED"
     * condition anywhere in the cause chain.
     */
    public static boolean isReactorStopped(Throwable t) {
        while (t != null) {
            String msg = t.getMessage();
            if (msg != null && msg.contains("I/O reactor") && msg.contains("STOPPED"))
                return true;
            t = t.getCause();
        }
        return false;
    }
}
