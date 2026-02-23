package com.park.utmstack.config.saml;

import com.park.utmstack.domain.idp_provider.IdentityProviderConfig;
import lombok.extern.slf4j.Slf4j;
import org.springframework.security.saml2.provider.service.registration.RelyingPartyRegistration;

import java.time.Duration;
import java.util.*;
import java.util.concurrent.*;

/**
 * Responsible for loading multiple SAML providers concurrently.
 * Separates the concern of concurrent loading and resource management.
 */
@Slf4j
public class SamlProvidersLoader {

    private static final Duration GLOBAL_LOADING_TIMEOUT = Duration.ofSeconds(30);
    private static final int MAX_CONCURRENT_LOADS = 5;

    private final SamlRegistrationBuilder registrationBuilder;

    public SamlProvidersLoader(SamlRegistrationBuilder registrationBuilder) {
        this.registrationBuilder = registrationBuilder;
    }

    /**
     * Loads multiple providers concurrently with global timeout.
     * Returns a map of loaded registrations (some may be missing if they failed).
     *
     * @param activeProviders List of provider configurations to load
     * @return Map of provider type -> registration (only successful loads)
     */
    public Map<String, RelyingPartyRegistration> loadProvidersAsync(
            List<IdentityProviderConfig> activeProviders) {

        if (activeProviders.isEmpty()) {
            log.info("No active SAML providers found");
            return new ConcurrentHashMap<>();
        }

        log.info("Starting async load of {} SAML provider(s)...", activeProviders.size());

        ExecutorService executor = null;
        try {
            executor = createExecutor(activeProviders.size());
            final ExecutorService finalExecutor = executor;
            Map<String, RelyingPartyRegistration> registrations = new ConcurrentHashMap<>();

            List<CompletableFuture<Void>> futures = activeProviders.stream()
                    .map(entity -> loadProviderAsync(entity, registrations, finalExecutor))
                    .toList();

            waitForAllLoads(futures);
            logLoadingResults(activeProviders.size(), registrations.size());

            return registrations;

        } finally {
            shutdownExecutorGracefully(executor);
        }
    }

    /**
     * Creates a thread pool executor for concurrent provider loading.
     */
    private ExecutorService createExecutor(int providerCount) {
        int poolSize = Math.min(MAX_CONCURRENT_LOADS, providerCount);
        return Executors.newFixedThreadPool(
                poolSize,
                r -> {
                    Thread t = new Thread(r);
                    t.setName("saml-provider-loader");
                    t.setDaemon(true);
                    t.setUncaughtExceptionHandler((thread, throwable) ->
                            log.error("Uncaught exception in provider loading thread: {}",
                                    throwable.getMessage(), throwable)
                    );
                    return t;
                }
        );
    }

    /**
     * Loads a single provider asynchronously.
     */
    private CompletableFuture<Void> loadProviderAsync(
            IdentityProviderConfig entity,
            Map<String, RelyingPartyRegistration> registrations,
            ExecutorService executor) {

        return CompletableFuture.runAsync(() -> {
            try {
                RelyingPartyRegistration registration = registrationBuilder.buildRegistration(entity);

                if (registration != null) {
                    registrations.put(entity.getProviderType().name().toLowerCase(), registration);
                    log.info("Successfully loaded SAML provider: {} (type: {})",
                            entity.getName(), entity.getProviderType());
                } else {
                    log.warn("SAML provider '{}' (type: {}) skipped - unable to load registration. " +
                                    "SSO will not be available for this provider type.",
                            entity.getName(), entity.getProviderType());
                }
            } catch (Exception e) {
                log.error("Unexpected error loading SAML provider '{}': {}",
                        entity.getName(), e.getMessage(), e);
            }
        }, executor);
    }

    /**
     * Waits for all provider loads to complete with global timeout.
     */
    private void waitForAllLoads(List<CompletableFuture<Void>> futures) {
        try {
            CompletableFuture.allOf(futures.toArray(new CompletableFuture[0]))
                    .get(GLOBAL_LOADING_TIMEOUT.getSeconds(), TimeUnit.SECONDS);
        } catch (TimeoutException e) {
            log.warn("Provider loading exceeded global timeout of {}s",
                    GLOBAL_LOADING_TIMEOUT.getSeconds());
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            log.error("Provider loading was interrupted");
        } catch (ExecutionException e) {
            log.error("Error during provider loading: {}", e.getMessage());
        }
    }

    /**
     * Logs summary of loading results.
     */
    private void logLoadingResults(int totalProviders, int loadedCount) {
        int failedCount = totalProviders - loadedCount;
        log.info("SAML provider loading completed: {} loaded, {} failed, {} total",
                loadedCount, failedCount, totalProviders);
    }

    /**
     * Safely shuts down the executor with proper timeout handling.
     */
    private void shutdownExecutorGracefully(ExecutorService executor) {
        if (executor == null) {
            return;
        }

        try {
            executor.shutdown();
            if (!executor.awaitTermination(5, TimeUnit.SECONDS)) {
                log.warn("Executor did not terminate in time, forcing shutdown");
                List<Runnable> droppedTasks = executor.shutdownNow();
                if (!droppedTasks.isEmpty()) {
                    log.warn("Dropped {} pending tasks during forced shutdown", droppedTasks.size());
                }

                if (!executor.awaitTermination(2, TimeUnit.SECONDS)) {
                    log.error("Executor did not terminate even after forced shutdown - potential thread leak");
                }
            }
        } catch (InterruptedException e) {
            log.error("Interrupted while waiting for executor termination");
            executor.shutdownNow();
            Thread.currentThread().interrupt();
        }
    }
}

