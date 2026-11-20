package com.park.utmstack.service;

import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.correlation.config.UtmDataTypes;
import com.park.utmstack.domain.correlation.rules.UtmCorrelationRules;
import com.park.utmstack.domain.logstash_filter.UtmLogstashFilter;
import com.park.utmstack.repository.correlation.config.UtmDataTypesRepository;
import com.park.utmstack.repository.correlation.rules.UtmCorrelationRulesRepository;
import com.park.utmstack.repository.logstash_filter.UtmLogstashFilterRepository;
import com.park.utmstack.service.correlation.rules.UtmCorrelationRulesService;
import com.park.utmstack.service.dto.correlation.AdversaryType;
import com.park.utmstack.service.dto.correlation.UtmCorrelationRulesDTO;
import com.park.utmstack.service.dto.correlation.UtmCorrelationRulesMapper;
import com.park.utmstack.service.dto.correlation.RuleYaml;
import com.park.utmstack.service.logstash_filter.UtmLogstashFilterService;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.CommandLineRunner;
import org.springframework.stereotype.Service;
import org.springframework.transaction.PlatformTransactionManager;
import org.springframework.transaction.support.TransactionTemplate;
import org.yaml.snakeyaml.LoaderOptions;
import org.yaml.snakeyaml.Yaml;
import org.yaml.snakeyaml.constructor.SafeConstructor;

import jakarta.validation.ConstraintViolation;
import jakarta.validation.Validator;

import java.io.IOException;
import java.io.UncheckedIOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.time.Instant;
import java.util.*;
import java.util.concurrent.atomic.AtomicBoolean;
import java.util.stream.Collectors;
import java.util.stream.Stream;

@Service
public class DefinitionSyncService implements CommandLineRunner {

    private final Logger log = LoggerFactory.getLogger(DefinitionSyncService.class);

    private final UtmLogstashFilterRepository filterRepository;
    private final UtmCorrelationRulesRepository rulesRepository;
    private final UtmDataTypesRepository dataTypesRepository;
    private final UtmCorrelationRulesService rulesService;
    private final UtmCorrelationRulesMapper rulesMapper;
    private final UtmLogstashFilterService filterService;
    private final Validator validator;
    private final PlatformTransactionManager transactionManager;

    private final TransactionTemplate txTemplate;

    public DefinitionSyncService(UtmLogstashFilterRepository filterRepository,
                                 UtmCorrelationRulesRepository rulesRepository,
                                 UtmDataTypesRepository dataTypesRepository,
                                 UtmCorrelationRulesService rulesService,
                                 UtmCorrelationRulesMapper rulesMapper,
                                 UtmLogstashFilterService filterService,
                                 Validator validator,
                                 PlatformTransactionManager transactionManager) {
        this.filterRepository = filterRepository;
        this.rulesRepository = rulesRepository;
        this.dataTypesRepository = dataTypesRepository;
        this.rulesService = rulesService;
        this.rulesMapper = rulesMapper;
        this.filterService = filterService;
        this.validator = validator;
        this.transactionManager = transactionManager;
        this.txTemplate = new TransactionTemplate(transactionManager);
    }

    /**
     * YAML parser that NEVER instantiates arbitrary Java objects. A bare
     * {@code new Yaml()} uses the unsafe default {@code Constructor}, which can
     * build any Java class from a {@code !!java...} tag — the core of
     * CVE-2022-1471 (snakeyaml RCE). These files are trusted (baked in by the
     * installer) so the unsafe constructor was never reachable in practice, but
     * SafeConstructor removes the risk entirely and is the CVE's recommended
     * mitigation. SafeConstructor handles plain maps/lists/scalars identically,
     * so parsing behavior is unchanged (no Java type tags or YAML dates are used
     * in the filter/rule files).
     */
    private static Yaml safeYaml() {
        return new Yaml(new SafeConstructor(new LoaderOptions()));
    }

    /**
     * Binds a plain YAML-parsed map (from {@link #safeYaml()} — only Map/List/scalars,
     * no Java type tags) into the {@link RuleYaml} DTO. Under snakeyaml 2.x (Boot 4)
     * {@code SafeConstructor.loadAs} refuses to construct arbitrary POJOs, so the bean
     * mapping is delegated to Jackson, which only sets properties on the concrete target
     * type — no arbitrary-class instantiation, no deserialization risk.
     */
    private static final ObjectMapper RULE_MAPPER = new ObjectMapper()
            .configure(com.fasterxml.jackson.databind.DeserializationFeature.FAIL_ON_UNKNOWN_PROPERTIES, false);

    private RuleYaml mapToRuleYaml(Map<String, Object> map) {
        if (map.containsKey("correlation") && !map.containsKey("afterEvents")) {
            map.put("afterEvents", map.remove("correlation"));
        }
        return RULE_MAPPER.convertValue(map, RuleYaml.class);
    }

    /**
     * Runs {@code work} inside its own short, self-contained transaction and isolates
     * the outcome: on success the transaction commits; on any failure it rolls back at
     * this boundary and the exception is logged and swallowed.
     *
     * <p>Each definition file is processed this way (see {@link #syncFilters()} and
     * {@link #syncRules()}) so a single bad file can never poison the surrounding boot
     * sequence. This replaces the old design where the whole sync ran in one shared
     * {@code @Transactional} boot transaction: if any file's write threw (even though a
     * per-file try/catch swallowed the exception), the shared transaction was left
     * marked rollback-only and the final commit threw
     * {@code UnexpectedRollbackException}, failing application startup into a crash
     * loop. With per-file transactions a failed file rolls back only its own work.
     *
     * <p>Running each file inside a transaction also keeps a Hibernate session open for
     * the whole file, which the lazy association reads (e.g.
     * {@code UtmDataTypes.getModule()}) require.
     */
    private void syncUnit(Path path, Runnable work, AtomicBoolean anyFailed) {
        try {
            txTemplate.executeWithoutResult(status -> work.run());
        } catch (Exception e) {
            anyFailed.set(true);
            log.error("Error syncing definition file {}: {}", path, e.getMessage(), e);
        }
    }

    @Override
    public void run(String... args) {
        log.info("Starting definition sync from filesystem... ---");
        try {
            AtomicBoolean filtersFailed = new AtomicBoolean(false);
            AtomicBoolean rulesFailed = new AtomicBoolean(false);

            Set<Long> filesystemFilters = syncFilters(filtersFailed);
            Set<String> filesystemRules = syncRules(rulesFailed);

            // Orphan cleanup is DESTRUCTIVE: it deletes any system row whose id/name is
            // not in the just-synced set. It must therefore only run when the scan of
            // that category is COMPLETE. If any file failed to process, the found set is
            // partial and a live row could look orphaned and be deleted. Skip cleanup on
            // a partial view (the safe direction: leave a stale row in place) and log it.
            if (filtersFailed.get()) {
                log.warn("Skipping filter orphan cleanup: at least one filter definition "
                        + "file failed to sync, so the discovered-filter set is incomplete.");
            } else {
                cleanupOrphanedFilters(filesystemFilters);
            }

            if (rulesFailed.get()) {
                log.warn("Skipping rule orphan cleanup: at least one rule definition "
                        + "file failed to sync, so the discovered-rule set is incomplete.");
            } else {
                cleanupOrphanedRules(filesystemRules);
            }

            log.info("Definition sync completed successfully. ---");
        } catch (Exception e) {
            // Never let a top-level sync failure fail application startup. The sync is
            // best-effort and re-runs on every start; each file already committed or
            // rolled back in its own transaction, so this catch is a last-resort guard
            // against the boot being taken down by a definition-sync error.
            log.error("CRITICAL: Definition sync failed. Reason: {} ---", e.getMessage(), e);
        }
    }

    private Set<Long> syncFilters(AtomicBoolean anyFailed) {
        Set<Long> foundFilters = new HashSet<>();
        Path filtersPath = Paths.get(".", Constants.APP_FILTER_DEFINITIONS);
        if (!Files.exists(filtersPath) || !Files.isDirectory(filtersPath)) {
            log.warn("Filters directory not found: {}", Constants.APP_FILTER_DEFINITIONS);
            return foundFilters;
        }

        try (Stream<Path> paths = Files.walk(filtersPath)) {
            paths.filter(path -> Files.isRegularFile(path) && isYamlFile(path))
                 .forEach(path -> syncUnit(path, () -> processFilterFile(path, foundFilters, anyFailed), anyFailed));
        } catch (IOException e) {
            anyFailed.set(true);
            log.error("Error listing filters directory: {}", e.getMessage());
        }
        return foundFilters;
    }

    private void processFilterFile(Path path, Set<Long> foundFilters, AtomicBoolean anyFailed) {
        String content;
        try {
            content = Files.readString(path);
        } catch (IOException e) {
            throw new UncheckedIOException("reading filter file", e);
        }
        if (content.isBlank()) {
            // A blank file cannot register the identity of any existing system row it
            // may correspond to; mark the category incomplete so cleanup is skipped
            // rather than delete a live row.
            anyFailed.set(true);
            log.warn("Skipping blank filter file: {}", path);
            return;
        }

        Yaml yaml = safeYaml();

        String dataTypeStr = extractFirstDataType(yaml.load(content));
        if (dataTypeStr == null) {
            // Same as above: no resolvable identity, so do not allow orphan cleanup.
            anyFailed.set(true);
            log.warn("Skipping filter file without dataType: {}", path);
            return;
        }

        log.info("found dataType: {}", dataTypeStr);

        Optional<UtmDataTypes> dataTypeEntity = dataTypesRepository.findOneByDataType(dataTypeStr.toLowerCase());

        String moduleName = null;
        if ("generic".equalsIgnoreCase(dataTypeStr)) {
            moduleName = "GENERIC";
        } else if (dataTypeEntity.isPresent() && dataTypeEntity.get().getModule() != null) {
            moduleName = dataTypeEntity.get().getModule().getModuleName().toString();
        }

        if (moduleName == null) {
            // Unknown module means this file cannot be mapped to a row; mark the
            // category incomplete so cleanup does not delete a live system row that
            // happens to correspond to it.
            anyFailed.set(true);
            log.error("module name for filter: {} with dataType: {} not found, ignoring...", path, dataTypeStr);
            return;
        }

        Optional<UtmLogstashFilter> filterOpt = filterRepository.findFirstByLogstashFilterAndSystemOwnerIsTrue(content);

        if (filterOpt.isPresent()) {
            UtmLogstashFilter filter = filterOpt.get();
            foundFilters.add(filter.getId());
            if (!content.equals(filter.getLogstashFilter())) {
                log.info("Updating existing filter for module: {}", moduleName);
                filter.setLogstashFilter(content);
                filter.setUpdatedAt(Instant.now());
                filterService.save(filter);
            }
        } else {
            UtmLogstashFilter filter = new UtmLogstashFilter();
            // Do NOT pre-set the id. UtmLogstashFilter implements Persistable, so save()
            // on a new instance routes to persist(), and Hibernate 7 rejects a pre-set
            // (non-unsaved) id with "Detached entity passed to persist". Let
            // CustomIdentityGenerator assign the id from the sequence at persist time —
            // the exact path the working API create uses. System vs user ownership is
            // tracked by system_owner, never by the id range, so the id partition
            // convention is not load-bearing for filters.
            filter.setModuleName(moduleName);
            filter.setFilterName(moduleName + " Filter");
            filter.setLogstashFilter(content);
            filter.setSystemOwner(true);
            filter.setActive(true);
            filter.setUpdatedAt(Instant.now());

            if (dataTypeEntity.isPresent()) {
                filter.setDatatype(dataTypeEntity.get());
            }

            log.info("Creating filter from file {} for module: {} and dataType {}", path, moduleName, dataTypeStr);
            filterService.save(filter);
            // The id is assigned by the generator at persist time. Verify it was
            // actually assigned before whitelisting the row: if it were null/absent
            // the cleanup would see the new row as an orphan and delete it.
            if (filter.getId() == null) {
                anyFailed.set(true);
                log.error("Created filter from file {} has no id after save; skipping "
                        + "filter orphan cleanup to avoid deleting it.", path);
                return;
            }
            // Capture it now (post save) so the orphaned-filter cleanup does not delete
            // the row we just created.
            foundFilters.add(filter.getId());
        }
    }

    private Set<String> syncRules(AtomicBoolean anyFailed) {
        Set<String> foundRules = new HashSet<>();
        Path rulesPath = Paths.get(".", Constants.APP_RULE_DEFINITIONS);
        if (!Files.exists(rulesPath) || !Files.isDirectory(rulesPath)) {
            log.warn("Rules directory not found: {}", Constants.APP_RULE_DEFINITIONS);
            return foundRules;
        }

        try (Stream<Path> paths = Files.walk(rulesPath)) {
            paths.filter(path -> Files.isRegularFile(path) && isYamlFile(path))
                 .forEach(path -> syncUnit(path, () -> processRuleFile(path, foundRules, anyFailed), anyFailed));
        } catch (IOException e) {
            anyFailed.set(true);
            log.error("Error listing rules directory: {}", e.getMessage());
        }
        return foundRules;
    }

    private void processRuleFile(Path path, Set<String> foundRules, AtomicBoolean anyFailed) {
        String content;
        try {
            content = Files.readString(path);
        } catch (IOException e) {
            throw new UncheckedIOException("reading rule file", e);
        }
        if (content.isBlank()) {
            // A blank rule file cannot register the identities of any rules it
            // previously held; mark the category incomplete so cleanup is skipped
            // rather than deleting live system rows.
            anyFailed.set(true);
            log.warn("Skipping blank rule file: {}", path);
            return;
        }

        Yaml yaml = safeYaml();
        Object yamlObj = yaml.load(content);
        List<Map<String, Object>> rawRules = new ArrayList<>();

        if (yamlObj instanceof List) {
            List<?> list = (List<?>) yamlObj;
            if (list.isEmpty()) {
                // An empty list maps to no rule identities at all.
                anyFailed.set(true);
                log.warn("Skipping rule file with an empty rule list: {}", path);
                return;
            }
            // Per-element type check: a list whose elements are not maps (e.g. a list of
            // scalars) must not abort the file with a ClassCastException — skip those
            // entries instead so the remaining valid rules still sync.
            for (Object element : list) {
                if (element instanceof Map) {
                    @SuppressWarnings("unchecked")
                    Map<String, Object> ruleMap = (Map<String, Object>) element;
                    rawRules.add(ruleMap);
                } else {
                    // Not a rule we can map to an identity: mark the category
                    // incomplete so cleanup does not delete a live row.
                    anyFailed.set(true);
                    log.warn("Skipping non-map rule entry in file: {}", path);
                }
            }
        } else if (yamlObj instanceof Map) {
            @SuppressWarnings("unchecked")
            Map<String, Object> ruleMap = (Map<String, Object>) yamlObj;
            rawRules.add(ruleMap);
        } else {
            // null (unparseable / comment-only document), a scalar, or any other
            // unsupported YAML root: it cannot be mapped to rule identities, so mark
            // the category incomplete rather than let cleanup delete live rows.
            anyFailed.set(true);
            log.warn("Skipping rule file with unsupported YAML root: {}", path);
            return;
        }

        for (Map<String, Object> rawRule : rawRules) {
            RuleYaml ruleYaml = mapToRuleYaml(rawRule);
            if (ruleYaml == null || ruleYaml.getName() == null) {
                // Cannot map this entry to a rule identity; mark the category
                // incomplete so cleanup does not delete a matching live system row.
                anyFailed.set(true);
                log.warn("Skipping invalid rule in file: {}", path);
                continue;
            }

            Optional<UtmCorrelationRules> ruleOpt = rulesRepository.findOneByRuleNameAndSystemOwner(ruleYaml.getName(), true);
            foundRules.add(ruleYaml.getName());
            UtmCorrelationRulesDTO ruleDto = new UtmCorrelationRulesDTO();

            if (ruleOpt.isPresent()) {
                ruleDto.setId(ruleOpt.get().getId());
            } else if (ruleYaml.getId() != null) {
                ruleDto.setId(ruleYaml.getId());
            } else {
                ruleDto.setId(rulesService.getSystemSequenceNextValue());
            }

            ruleDto.setName(ruleYaml.getName());
            ruleDto.setCategory(ruleYaml.getCategory());
            ruleDto.setTechnique(ruleYaml.getTechnique());
            ruleDto.setAdversary(ruleYaml.getAdversary() != null ? ruleYaml.getAdversary() : AdversaryType.origin);
            ruleDto.setDescription(ruleYaml.getDescription());
            ruleDto.setReferences(ruleYaml.getReferences());
            ruleDto.setDefinition(ruleYaml.getWhere());
            ruleDto.setGroupBy(ruleYaml.getGroupBy());
            ruleDto.setDeduplicateBy(ruleYaml.getDeduplicateBy());
            ruleDto.setAfterEvents(ruleYaml.getAfterEvents());
            ruleDto.setSystemOwner(true);
            ruleDto.setRuleActive(ruleOpt.isPresent() ? ruleOpt.get().getRuleActive() : false);

            if (ruleYaml.getImpact() != null) {
                ruleDto.setConfidentiality(ruleYaml.getImpact().getConfidentiality());
                ruleDto.setIntegrity(ruleYaml.getImpact().getIntegrity());
                ruleDto.setAvailability(ruleYaml.getImpact().getAvailability());
            }

            // Map dataTypes strings to UtmDataTypes entities
            if (ruleYaml.getDataTypes() != null) {
                Set<UtmDataTypes> dataTypes = ruleYaml.getDataTypes().stream()
                        .map(dtName -> dataTypesRepository.findOneByDataType(dtName.toLowerCase()))
                        .filter(Optional::isPresent)
                        .map(Optional::get)
                        .collect(Collectors.toSet());
                ruleDto.setDataTypes(dataTypes);
            }

            Set<ConstraintViolation<UtmCorrelationRulesDTO>> violations = validator.validate(ruleDto);
            if (!violations.isEmpty()) {
                log.error("Skipping invalid rule '{}' in file {}: {}",
                        ruleYaml.getName(), path,
                        violations.stream().map(ConstraintViolation::getMessage).collect(Collectors.joining("; ")));
                continue;
            }

            UtmCorrelationRules entity = rulesMapper.toEntity(ruleDto);
            if (ruleOpt.isPresent()) {
                try {
                    rulesService.updateRule(entity, true);
                } catch (Exception e) {
                    // updateRule declares a checked exception; rethrow unchecked so it
                    // propagates out of the per-file transaction and rolls back ONLY this
                    // file's work (the syncUnit boundary logs and swallows it).
                    throw new RuntimeException("updateRule failed for rule '" + ruleDto.getName() + "'", e);
                }
            } else {
                rulesService.save(entity, true);
            }
        }
    }

    /**
     * Extract the first dataType from a parsed filter file's pipeline structure.
     * Handles both list syntaxes:
     *   pipeline:
     *   - dataTypes:
     *     - value
     *   and flow syntax:
     *   pipeline:
     *   - dataTypes: [value]
     */
    @SuppressWarnings("unchecked")
    private String extractFirstDataType(Object yamlObj) {
        if (!(yamlObj instanceof Map)) {
            return null;
        }
        Object pipeline = ((Map<String, Object>) yamlObj).get("pipeline");
        if (!(pipeline instanceof List)) {
            return null;
        }
        for (Object entry : (List<?>) pipeline) {
            if (!(entry instanceof Map)) {
                continue;
            }
            Object dataTypes = ((Map<String, Object>) entry).get("dataTypes");
            if (dataTypes instanceof List) {
                List<?> list = (List<?>) dataTypes;
                if (!list.isEmpty() && list.get(0) != null) {
                    return list.get(0).toString();
                }
            } else if (dataTypes != null) {
                return dataTypes.toString();
            }
        }
        return null;
    }

    private void cleanupOrphanedFilters(Set<Long> currentFilterIds) {
        if (currentFilterIds.isEmpty()) return;
        List<UtmLogstashFilter> systemFilters = filterRepository.findAllBySystemOwnerIsTrue();
        for (UtmLogstashFilter filter : systemFilters) {
            if (currentFilterIds.contains(filter.getId())) {
                continue;
            }
            log.info("Deleting orphaned system filter: {}", filter.getModuleName());
            try {
                filterService.delete(filter.getId());
            } catch (Exception e) {
                log.error("Error deleting orphaned system filter {}: {}", filter.getId(), e.getMessage());
            }
        }
    }

    private void cleanupOrphanedRules(Set<String> currentFilesystemRules) {
        if (currentFilesystemRules.isEmpty()) return;

        List<UtmCorrelationRules> systemRules = rulesRepository.findAllBySystemOwnerIsTrue();
        for (UtmCorrelationRules rule : systemRules) {
            if (currentFilesystemRules.contains(rule.getRuleName())) {
                continue;
            }
            log.info("Deleting orphaned system rule: {}", rule.getRuleName());
            try {
                rulesService.deleteRule(rule.getId(), true);
            } catch (Exception e) {
                log.error("Error deleting orphaned system rule {}: {}", rule.getId(), e.getMessage());
            }
        }
    }

    private boolean isYamlFile(Path path) {
        String fileName = path.getFileName().toString().toLowerCase();
        return fileName.endsWith(".yaml") || fileName.endsWith(".yml");
    }

    private String getFileNameWithoutExtension(Path path) {
        String fileName = path.getFileName().toString();
        int lastDotIndex = fileName.lastIndexOf('.');
        return (lastDotIndex == -1) ? fileName : fileName.substring(0, lastDotIndex);
    }
}
