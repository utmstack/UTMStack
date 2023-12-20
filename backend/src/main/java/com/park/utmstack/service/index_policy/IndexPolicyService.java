package com.park.utmstack.service.index_policy;

import com.google.gson.Gson;
import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.index_pattern.enums.SystemIndexPattern;
import com.park.utmstack.domain.index_policy.*;
import com.park.utmstack.service.elasticsearch.OpensearchClientBuilder;
import com.park.utmstack.util.events.IndexPatternsReadyEvent;
import com.utmstack.opensearch_connector.enums.HttpMethod;
import okhttp3.Response;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.SpringApplication;
import org.springframework.context.ApplicationContext;
import org.springframework.context.event.EventListener;
import org.springframework.stereotype.Service;
import org.springframework.util.StringUtils;

import java.util.List;
import java.util.Map;
import java.util.Optional;

@Service
public class IndexPolicyService {
    private static final String CLASSNAME = "IndexPolicyService";
    private final Logger log = LoggerFactory.getLogger(IndexPolicyService.class);

    private final ApplicationContext applicationContext;
    private final String CURRENT_POLICY_ID = "utmstack_ism_policy";
    private final String SNAPSHOT_REPOSITORY_NAME = "utmstack_backups";
    private final OpensearchClientBuilder client;


    public IndexPolicyService(ApplicationContext applicationContext,
                              OpensearchClientBuilder client) {
        this.applicationContext = applicationContext;
        this.client = client;
    }

    @EventListener(IndexPatternsReadyEvent.class)
    public void init() throws Exception {
        final String ctx = CLASSNAME + ".init";
        try {
            // 1. Creating the UtmStack main index management policy if it does not exist
            createPolicy();

            // 2. Registering snapshots repository
            registerSnapshotRepository();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            System.exit(SpringApplication.exit(applicationContext, () -> -1));
        }
    }

    /**
     * Create the main UtmStack index management policy if it does not exist
     */
    private void createPolicy() {
        final String ctx = CLASSNAME + ".createPolicy";
        try {
            if (getPolicy().isPresent())
                return;
            IndexPolicy body = buildDefaultIndexPolicy();
            final String uri = String.format("_plugins/_ism/policies/%1$s", CURRENT_POLICY_ID);
            try (Response rs = client.getClient().executeHttpRequest(uri, null, body, HttpMethod.PUT)) {
                if (!rs.isSuccessful())
                    throw new RuntimeException(rs.body() != null ? rs.body().string() : rs.toString());
                if (rs.body() != null)
                    log.info(ctx + ": " + rs.body().string());

                // Adds a policy to an index. This operation does not change the policy if the index already has one.
                addPolicy(Constants.SYS_INDEX_PATTERN.get(SystemIndexPattern.LOGS));
                addPolicy(Constants.SYS_INDEX_PATTERN.get(SystemIndexPattern.ALERTS));
            }
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    /**
     * Gets the index policy
     *
     * @return A {@link IndexPolicy} object
     */
    public Optional<IndexPolicy> getPolicy() {
        final String ctx = CLASSNAME + ".getPolicy";
        try {
            final String url = String.format("_plugins/_ism/policies/%1$s", CURRENT_POLICY_ID);
            try (Response rs = client.getClient().executeHttpRequest(url, null, null, HttpMethod.GET)) {
                if (rs.code() == 404)
                    return Optional.empty();
                if (!rs.isSuccessful())
                    throw new Exception(rs.body() != null ? rs.body().string() : rs.toString());
                if (rs.body() != null) {
                    String body = rs.body().string();
                    if (!StringUtils.hasText(body))
                        return Optional.empty();
                    return Optional.ofNullable(new Gson().fromJson(body, IndexPolicy.class));
                }
            }
            return Optional.empty();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            throw new RuntimeException(msg);
        }
    }

    /**
     * Verifies if there are any change to apply to the policy
     *
     * @param indexPolicy Current index policy
     * @param settings    Configurations to apply
     */
    private boolean anyChange(IndexPolicy indexPolicy, PolicySettings settings) {
        final String ctx = CLASSNAME + ".detectPolicyChanges";
        try {
            // Detecting any change over backup configuration
            Optional<State> ingest = indexPolicy.getPolicy().getStates().stream()
                .filter(s -> s.getName().equals(Constants.STATE_INGEST)).findFirst();
            if (ingest.isEmpty())
                throw new Exception("State ingest is missing");
            boolean backupActive = ingest.get().getTransitions().stream()
                .map(Transition::getStateName).anyMatch(n -> n.equals(Constants.STATE_BACKUP));
            if (backupActive != settings.isSnapshotActive())
                return true;

            // Detecting any change over retention time configuration
            Optional<State> open = indexPolicy.getPolicy().getStates().stream()
                .filter(s -> s.getName().equals(Constants.STATE_OPEN)).findFirst();
            if (open.isEmpty())
                throw new Exception("State open is missing");
            return !open.get().getTransitions().get(0).getConditions().getMinIndexAge()
                .equals(settings.getDeleteAfter());
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    /**
     * Updates a policy. Use the seq_no and primary_term parameters to update an existing policy.
     * If these numbers don’t match the existing policy or the policy doesn’t exist, ISM throws an error.
     * It’s possible that the policy currently applied to your index isn’t the most up-to-date policy available.
     *
     * @param settings Configurations to apply
     */
    public void updatePolicy(PolicySettings settings) {
        final String ctx = CLASSNAME + ".updatePolicy";
        try {
            IndexPolicy indexPolicy = getPolicy()
                .orElseThrow(() -> new Exception(String.format("Policy with id: %1$s not found", CURRENT_POLICY_ID)));

            if (!anyChange(indexPolicy, settings))
                return;

            List<State> states = indexPolicy.getPolicy().getStates();

            if (settings.isSnapshotActive()) {
                // ingest -> backup
                states.stream().filter(s -> s.getName().equals(Constants.STATE_INGEST)).findFirst()
                    .ifPresent(state -> {
                        Transition transition = state.getTransitions().get(0);
                        transition.setStateName(Constants.STATE_BACKUP);
                        transition.setConditions(null);
                    });

                // open -> safe_delete
                states.stream().filter(s -> s.getName().equals(Constants.STATE_OPEN)).findFirst()
                    .ifPresent(state -> {
                        Transition transition = state.getTransitions().get(0);
                        transition.setStateName(Constants.STATE_SAFE_DELETE);
                        transition.setConditions(new TransitionCondition(settings.getDeleteAfter()));
                    });
            } else {
                // ingest -> open
                states.stream().filter(s -> s.getName().equals(Constants.STATE_INGEST)).findFirst()
                    .ifPresent(state -> {
                        Transition transition = state.getTransitions().get(0);
                        transition.setStateName(Constants.STATE_OPEN);
                    });

                // open -> delete
                states.stream().filter(s -> s.getName().equals(Constants.STATE_OPEN)).findFirst()
                    .ifPresent(state -> {
                        Transition transition = state.getTransitions().get(0);
                        transition.setStateName(Constants.STATE_DELETE);
                        transition.setConditions(new TransitionCondition(settings.getDeleteAfter()));
                    });
            }

            final String uri = String.format("/_plugins/_ism/policies/%1$s", CURRENT_POLICY_ID);
            Map<String, String> params = Map.of("if_seq_no", String.valueOf(indexPolicy.getSeqNo()),
                "if_primary_term", String.valueOf(indexPolicy.getPrimaryTerm()));
            try (Response rs = client.getClient().executeHttpRequest(uri, params, new IndexPolicy(indexPolicy.getPolicy()), HttpMethod.PUT)) {
                if (!rs.isSuccessful())
                    throw new Exception(rs.body() != null ? rs.body().string() : rs.toString());
            }
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            throw new RuntimeException(msg);
        }
    }

    /**
     * Updates the managed index policy to a new policy (or to a new version of the policy).
     * You can use an index pattern to update multiple indexes at once
     *
     * @param snapshotActivated Define if action to configure index backups are active or not
     * @return A ${@link UpdateManagedIndexPolicyResponse} object
     */
    public UpdateManagedIndexPolicyResponse updateManagedIndexPolicy(boolean snapshotActivated) {
        final String ctx = CLASSNAME + ".updateManagedIndexPolicy";
        try {
            UpdateManagedIndexPolicyResponse result = new UpdateManagedIndexPolicyResponse();

            if (snapshotActivated) {
                UpdateManagedIndexPolicyResponse ingestToIngestForLogs = updateManagedIndex(Constants.SYS_INDEX_PATTERN.get(SystemIndexPattern.LOGS),
                    Constants.STATE_INGEST, Constants.STATE_INGEST);
                UpdateManagedIndexPolicyResponse openToBackupForLogs = updateManagedIndex(Constants.SYS_INDEX_PATTERN.get(SystemIndexPattern.LOGS),
                    Constants.STATE_OPEN, Constants.STATE_BACKUP);
                UpdateManagedIndexPolicyResponse ingestToIngestForAlerts = updateManagedIndex(Constants.SYS_INDEX_PATTERN.get(SystemIndexPattern.ALERTS),
                    Constants.STATE_INGEST, Constants.STATE_INGEST);
                UpdateManagedIndexPolicyResponse openToBackupForAlerts = updateManagedIndex(Constants.SYS_INDEX_PATTERN.get(SystemIndexPattern.ALERTS),
                    Constants.STATE_OPEN, Constants.STATE_BACKUP);

                int updatedIndices = ingestToIngestForLogs.getUpdatedIndices() + openToBackupForLogs.getUpdatedIndices() + ingestToIngestForAlerts.getUpdatedIndices()
                    + openToBackupForAlerts.getUpdatedIndices();

                result.setUpdatedIndices(updatedIndices);
                result.getFailedIndices().addAll(ingestToIngestForLogs.getFailedIndices());
                result.getFailedIndices().addAll(openToBackupForLogs.getFailedIndices());
                result.getFailedIndices().addAll(ingestToIngestForAlerts.getFailedIndices());
                result.getFailedIndices().addAll(openToBackupForAlerts.getFailedIndices());
            } else {
                UpdateManagedIndexPolicyResponse ingestToIngestForLogs = updateManagedIndex(Constants.SYS_INDEX_PATTERN.get(SystemIndexPattern.LOGS),
                    Constants.STATE_INGEST, Constants.STATE_INGEST);
                UpdateManagedIndexPolicyResponse backupToOpenForLogs = updateManagedIndex(Constants.SYS_INDEX_PATTERN.get(SystemIndexPattern.LOGS),
                    Constants.STATE_BACKUP, Constants.STATE_OPEN);
                UpdateManagedIndexPolicyResponse openToOpenForLogs = updateManagedIndex(Constants.SYS_INDEX_PATTERN.get(SystemIndexPattern.LOGS),
                    Constants.STATE_OPEN, Constants.STATE_OPEN);
                UpdateManagedIndexPolicyResponse ingestToIngestForAlerts = updateManagedIndex(Constants.SYS_INDEX_PATTERN.get(SystemIndexPattern.ALERTS),
                    Constants.STATE_INGEST, Constants.STATE_INGEST);
                UpdateManagedIndexPolicyResponse backupToOpenForAlerts = updateManagedIndex(Constants.SYS_INDEX_PATTERN.get(SystemIndexPattern.ALERTS),
                    Constants.STATE_BACKUP, Constants.STATE_OPEN);
                UpdateManagedIndexPolicyResponse openToOpenForAlerts = updateManagedIndex(Constants.SYS_INDEX_PATTERN.get(SystemIndexPattern.ALERTS),
                    Constants.STATE_OPEN, Constants.STATE_OPEN);

                int updatedIndices = ingestToIngestForLogs.getUpdatedIndices() + backupToOpenForLogs.getUpdatedIndices() + openToOpenForLogs.getUpdatedIndices()
                    + ingestToIngestForAlerts.getUpdatedIndices() + backupToOpenForAlerts.getUpdatedIndices() + openToOpenForAlerts.getUpdatedIndices();

                result.setUpdatedIndices(updatedIndices);
                result.getFailedIndices().addAll(ingestToIngestForLogs.getFailedIndices());
                result.getFailedIndices().addAll(backupToOpenForLogs.getFailedIndices());
                result.getFailedIndices().addAll(openToOpenForLogs.getFailedIndices());
                result.getFailedIndices().addAll(ingestToIngestForAlerts.getFailedIndices());
                result.getFailedIndices().addAll(backupToOpenForAlerts.getFailedIndices());
                result.getFailedIndices().addAll(openToOpenForAlerts.getFailedIndices());
            }
            result.setFailures(!result.getFailedIndices().isEmpty());
            return result;
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            throw new RuntimeException(msg);
        }
    }

    private UpdateManagedIndexPolicyResponse updateManagedIndex(String index, String sourceState, String destinationState) {
        final String ctx = CLASSNAME + ".updateManagedIndexPolicy";
        try {
            final String uri = String.format("_plugins/_ism/change_policy/%1$s", index);
            UpdateManagedIndexPolicyResponse result = new UpdateManagedIndexPolicyResponse();

            UpdateManagedIndexPolicyConfiguration config = UpdateManagedIndexPolicyConfiguration.builder()
                .withPolicyId(CURRENT_POLICY_ID).sourceStates(sourceState).destinationState(destinationState).build();

            try (Response rs = client.getClient().executeHttpRequest(uri, null, config, HttpMethod.POST)) {
                if (rs.isSuccessful() && rs.body() != null) {
                    UpdateManagedIndexPolicyResponse body = new Gson().fromJson(rs.body().string(), UpdateManagedIndexPolicyResponse.class);
                    result.setUpdatedIndices(result.getUpdatedIndices() + body.getUpdatedIndices());
                    result.getFailedIndices().addAll(body.getFailedIndices());
                }
                return result;
            }
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    /**
     * Check if the UtmStack snapshots repository exist.
     *
     * @return True if UtmStack snapshots repository exist, False otherwise
     * @throws Exception In case of any error
     */
    private boolean existSnapshotRepository() throws Exception {
        final String ctx = CLASSNAME + ".existSnapshotRepository";
        try {
            final String uri = String.format("/_snapshot/%1$s", SNAPSHOT_REPOSITORY_NAME);
            try (Response rs = client.getClient().executeHttpRequest(uri, null, null, HttpMethod.GET)) {
                if (rs.code() == 404)
                    return false;
                if (!rs.isSuccessful())
                    throw new Exception(rs.body() != null ? rs.body().string() : rs.toString());
                return true;
            }
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Register the UtmStack snapshots repository if it does not exist
     */
    private void registerSnapshotRepository() {
        final String ctx = CLASSNAME + ".registerSnapshotRepository";
        try {
            if (existSnapshotRepository())
                return;
            final String uri = String.format("/_snapshot/%1$s", SNAPSHOT_REPOSITORY_NAME);
            try (Response rs = client.getClient().executeHttpRequest(uri, null, new RegisterRepositoryRequest(), HttpMethod.PUT)) {
                if (!rs.isSuccessful())
                    throw new Exception(rs.body() != null ? rs.body().string() : rs.toString());
            }
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Adds a policy to an index. This operation does not change the policy if the index already has one.
     *
     * @param index Specify the index or the pattern for which you want to add the policy
     */
    private void addPolicy(String index) {
        final String ctx = CLASSNAME + ".addPolicy";
        try {
            final String uri = String.format("/_plugins/_ism/add/%1$s", index);
            try (Response rs = client.getClient().executeHttpRequest(uri, null, new AddPolicyRequest(CURRENT_POLICY_ID), HttpMethod.POST)) {
                if (!rs.isSuccessful())
                    throw new Exception(rs.body() != null ? rs.body().string() : rs.toString());
                if (rs.body() != null)
                    log.info(ctx + ": " + rs.body().string());
            }
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    private IndexPolicy buildDefaultIndexPolicy() {
        final String ctx = CLASSNAME + ".createPolicy";
        try {
            return IndexPolicy.builder()

                .withPolicy(Policy.builder()
                    .withDefaultState(Constants.STATE_INGEST)
                    .withDescription("UtmStack main index lifecycle")

                    .withIsmTemplate(IsmTemplate.builder()
                        .withIndexPattern(Constants.SYS_INDEX_PATTERN.get(SystemIndexPattern.LOGS))
                        .withIndexPattern(Constants.SYS_INDEX_PATTERN.get(SystemIndexPattern.ALERTS))
                        .withPriority(1)
                        .build())

                    .withState(State.builder()
                        .withName(Constants.STATE_INGEST)
                        .transitionTo(Transition.builder()
                            .stateName(Constants.STATE_OPEN)
                            .ifMinIndexAgeIs("24h")
                            .build())
                        .build())

                    .withState(State.builder()
                        .withName(Constants.STATE_OPEN)
                        .transitionTo(Transition.builder()
                            .stateName(Constants.STATE_DELETE)
                            .ifMinIndexAgeIs("30d")
                            .build())
                        .build())

                    .withState(State.builder()
                        .withName(Constants.STATE_BACKUP)
                        .executeAction(Action.builder()
                            .snapshot(new ActionSnapshot(SNAPSHOT_REPOSITORY_NAME, "utmstack_snapshot"))
                            .build())
                        .transitionTo(Transition.builder()
                            .stateName(Constants.STATE_OPEN)
                            .build())
                        .build())

                    .withState(State.builder()
                        .withName(Constants.STATE_DELETE)
                        .executeAction(Action.builder()
                            .delete(new ActionDelete())
                            .build())
                        .build())

                    .withState(State.builder()
                        .withName(Constants.STATE_SAFE_DELETE)
                        .executeAction(Action.builder()
                            .snapshot(new ActionSnapshot(SNAPSHOT_REPOSITORY_NAME, "utmstack_snapshot"))
                            .build())
                        .executeAction(Action.builder()
                            .delete(new ActionDelete())
                            .build())
                        .build())
                    .build())

                .build();
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }
}
