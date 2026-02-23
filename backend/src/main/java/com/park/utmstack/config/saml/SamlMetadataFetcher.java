package com.park.utmstack.config.saml;

import com.park.utmstack.domain.idp_provider.IdentityProviderConfig;
import lombok.extern.slf4j.Slf4j;
import org.springframework.security.saml2.provider.service.registration.RelyingPartyRegistration;
import org.springframework.security.saml2.provider.service.registration.RelyingPartyRegistrations;
import org.springframework.stereotype.Component;

import java.time.Duration;
import java.util.concurrent.*;

/**
 * Responsible for fetching SAML metadata with timeout handling.
 * Separates the concern of async metadata fetching from registration building.
 */
@Slf4j
public class SamlMetadataFetcher {

    private static final Duration METADATA_FETCH_TIMEOUT = Duration.ofSeconds(10);

    /**
     * Fetches SAML metadata with timeout protection.
     * Returns null if timeout, error, or interruption occurs.
     * Logs detailed error information instead of throwing exceptions.
     *
     * @param entity Provider configuration
     * @return RelyingPartyRegistration, or null if fetch fails
     */
    public RelyingPartyRegistration fetchMetadataWithTimeout(IdentityProviderConfig entity) {
        ExecutorService timeoutExecutor = null;
        Future<RelyingPartyRegistration> future = null;

        try {
            timeoutExecutor = createMetadataFetchExecutor(entity);

            future = CompletableFuture.supplyAsync(() -> {
                try {
                    return RelyingPartyRegistrations
                            .fromMetadataLocation(entity.getMetadataUrl())
                            .registrationId(entity.getName())
                            .build();
                } catch (Exception e) {
                    throw new CompletionException(e);
                }
            }, timeoutExecutor);

            return future.get(METADATA_FETCH_TIMEOUT.getSeconds(), TimeUnit.SECONDS);

        } catch (TimeoutException e) {
            handleTimeoutException(entity, future, e);
            return null;

        } catch (ExecutionException e) {
            handleExecutionException(entity, e);
            return null;

        } catch (InterruptedException e) {
            handleInterruptedException(entity, e);
            return null;

        } finally {
            cleanupExecutor(entity, timeoutExecutor);
        }
    }

    /**
     * Creates an executor for metadata fetching with proper naming and exception handling.
     */
    private ExecutorService createMetadataFetchExecutor(IdentityProviderConfig entity) {
        return Executors.newSingleThreadExecutor(r -> {
            Thread t = new Thread(r);
            t.setName("saml-metadata-fetch-" + entity.getName());
            t.setDaemon(true);
            t.setUncaughtExceptionHandler((thread, throwable) ->
                    log.error("Uncaught exception in SAML metadata fetch thread for {}: {}",
                            entity.getName(), throwable.getMessage(), throwable)
            );
            return t;
        });
    }

    /**
     * Handles timeout exception with detailed logging.
     */
    private void handleTimeoutException(IdentityProviderConfig entity, Future<?> future, TimeoutException e) {
        if (future != null) {
            future.cancel(true);
        }
        log.error(
                "SAML metadata fetch TIMEOUT: Provider='{}', Timeout={}s, MetadataUrl='{}'. " +
                        "This provider will not be available for SSO until it responds faster or the endpoint is fixed.",
                entity.getName(),
                METADATA_FETCH_TIMEOUT.getSeconds(),
                entity.getMetadataUrl(),
                e
        );
    }

    /**
     * Handles execution exception with root cause extraction.
     */
    private void handleExecutionException(IdentityProviderConfig entity, ExecutionException e) {
        Throwable rootCause = e.getCause() != null ? e.getCause() : e;
        log.error(
                "SAML metadata fetch FAILED: Provider='{}'. Root cause: {}. " +
                        "Error details: {}. This provider will not be available for SSO.",
                entity.getName(),
                rootCause.getClass().getSimpleName(),
                rootCause.getMessage(),
                rootCause
        );
    }

    /**
     * Handles interruption exception.
     */
    private void handleInterruptedException(IdentityProviderConfig entity, InterruptedException e) {
        Thread.currentThread().interrupt();
        log.error(
                "SAML metadata fetch INTERRUPTED: Provider='{}'. " +
                        "Current thread was interrupted. Thread status restored. " +
                        "This provider will not be available for SSO.",
                entity.getName(),
                e
        );
    }

    /**
     * Safely shuts down the executor and logs any issues.
     */
    private void cleanupExecutor(IdentityProviderConfig entity, ExecutorService executor) {
        if (executor != null) {
            try {
                executor.shutdownNow();

                if (!executor.awaitTermination(2, TimeUnit.SECONDS)) {
                    log.warn(
                            "Executor for SAML provider '{}' did not terminate cleanly within 2 seconds. " +
                                    "Potential thread leak detected.",
                            entity.getName()
                    );
                }
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
                log.error(
                        "Interrupted while waiting for executor shutdown for SAML provider '{}'. " +
                                "Thread status restored.",
                        entity.getName(),
                        e
                );
            }
        }
    }
}

