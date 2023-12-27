package com.utmstack.userauditor.service.elasticsearch;

import com.utmstack.opensearch_connector.OpenSearch;
import com.utmstack.opensearch_connector.enums.HttpScheme;
import lombok.Getter;
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
    @Getter
    private OpenSearch client;

    @Order(Ordered.HIGHEST_PRECEDENCE)
    @EventListener(ApplicationReadyEvent.class)
    public void init() throws Exception {
        final String ctx = CLASSNAME + ".init";
        try {
            String host = System.getenv("ELASTICSEARCH_HOST");
            Assert.hasText(host, "Environment variable ELASTICSEARCH_HOST is missing or his value is null or empty");

            String port = System.getenv("ELASTICSEARCH_PORT");
            Assert.hasText(host, "Environment variable ELASTICSEARCH_PORT is missing or his value is null or empty");

           client = OpenSearch.builder().withHost(host, Integer.parseInt(port), HttpScheme.http).build();

        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            throw new RuntimeException(msg);
        }
    }
}
