package com.utmstack.userauditor.service.elasticsearch;

import com.utmstack.opensearch_connector.OpenSearch;
import com.utmstack.opensearch_connector.enums.HttpScheme;
import lombok.Getter;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;
import org.springframework.util.Assert;

@Service
public class OpensearchClientBuilder {
    private static final String CLASSNAME = "OpensearchClientBuilder";
    private final Logger log = LoggerFactory.getLogger(OpensearchClientBuilder.class);
    @Getter
    private OpenSearch client;

    /**
     * The OpenSearch client is built eagerly in the constructor, NOT on
     * ApplicationReadyEvent. The @Scheduled findUsers() tick can fire before
     * ApplicationReadyEvent is published, and a null client there threw
     * "OpensearchClientBuilder.getClient() is null" on every restart. Because
     * OpensearchClientBuilder is a constructor dependency of
     * ElasticsearchService -> WindowsSource -> UserService, it is fully
     * constructed before the singleton UserService (which holds the
     * @Scheduled task) is created, so the client is guaranteed non-null by
     * the time the scheduler can run. Failing here also fails fast at startup
     * if the ELASTICSEARCH_* env vars are missing, same as before.
     */
    public OpensearchClientBuilder() {
        final String ctx = CLASSNAME + ".constructor";
        try {
            String host = System.getenv("ELASTICSEARCH_HOST");
            Assert.hasText(host, "Environment variable ELASTICSEARCH_HOST is missing or his value is null or empty");

            String port = System.getenv("ELASTICSEARCH_PORT");
            Assert.hasText(port, "Environment variable ELASTICSEARCH_PORT is missing or his value is null or empty");

            String user = System.getenv("ELASTICSEARCH_USER");
            Assert.hasText(user, "Environment variable ELASTICSEARCH_USER is missing or his value is null or empty");

            String password = System.getenv("ELASTICSEARCH_PASSWORD");
            Assert.hasText(password, "Environment variable ELASTICSEARCH_PASSWORD is missing or his value is null or empty");

            client = OpenSearch.builder()
                    .withHost(host, Integer.parseInt(port), HttpScheme.https)
                    .withCredentials(user, password)
                    .build();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            throw new RuntimeException(msg);
        }
    }
}
