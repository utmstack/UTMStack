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
import com.park.utmstack.service.logstash_filter.UtmLogstashFilterService;
import lombok.Data;
import lombok.RequiredArgsConstructor;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.CommandLineRunner;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.yaml.snakeyaml.Yaml;

import javax.annotation.PostConstruct;
import java.io.IOException;
import java.lang.Exception;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.time.Instant;
import java.util.*;
import java.util.stream.Collectors;
import java.util.stream.Stream;

@Service
@RequiredArgsConstructor
public class DefinitionSyncService implements CommandLineRunner {

    private final Logger log = LoggerFactory.getLogger(DefinitionSyncService.class);

    private final UtmLogstashFilterRepository filterRepository;
    private final UtmCorrelationRulesRepository rulesRepository;
    private final UtmDataTypesRepository dataTypesRepository;
    private final UtmCorrelationRulesService rulesService;
    private final UtmCorrelationRulesMapper rulesMapper;
    private final UtmLogstashFilterService filterService;

    @Override
    public void run(String... args) {
        log.info("Starting definition sync from filesystem... ---");
        try {
            Set<String> filesystemFilters = syncFilters();
            Set<String> filesystemRules = syncRules();

            cleanupOrphanedFilters(filesystemFilters);
            cleanupOrphanedRules(filesystemRules);

            log.info("Definition sync completed successfully. ---");
        } catch (Exception e) {
            log.error("CRITICAL: Definition sync failed. Reason: {} ---", e.getMessage(), e);
        }
    }

    private Set<String> syncFilters() {
        Set<String> foundModules = new HashSet<>();
        Path filtersPath = Paths.get(".",Constants.APP_FILTER_DEFINITIONS);
        if (!Files.exists(filtersPath) || !Files.isDirectory(filtersPath)) {
            log.warn("Filters directory not found: {}", Constants.APP_FILTER_DEFINITIONS);
            return foundModules;
        }

        try (Stream<Path> paths = Files.walk(filtersPath)) {
            paths.filter(path -> Files.isRegularFile(path) && isYamlFile(path)).forEach(path -> {
                String rawName = getFileNameWithoutExtension(path);
                String moduleName = rawName.toUpperCase().replace("-", "_");
                foundModules.add(moduleName);
                try {
                    String content = Files.readString(path);

                    Optional<UtmLogstashFilter> filterOpt = filterRepository.findOneByModuleName(moduleName);

                    if (filterOpt.isPresent()) {
                        UtmLogstashFilter filter = filterOpt.get();
                        if (!content.equals(filter.getLogstashFilter())) {
                            log.info("Updating existing filter for module: {}", moduleName);
                            filter.setLogstashFilter(content);
                            filter.setUpdatedAt(Instant.now());
                            filterService.save(filter, true);
                        }
                    } else {
                        UtmLogstashFilter filter = new UtmLogstashFilter();
                        filter.setModuleName(moduleName);
                        filter.setFilterName(moduleName + " Filter");
                        filter.setLogstashFilter(content);
                        filter.setSystemOwner(true);
                        filter.setActive(true);
                        filter.setUpdatedAt(Instant.now());

                        // Try to find a matching data type
                        Optional<UtmDataTypes> dataType = dataTypesRepository.findOneByDataType(moduleName.toLowerCase());
                        if (dataType.isPresent()) {
                            filter.setDatatype(dataType.get());
                        }

                        filterService.save(filter, true);
                    }
                } catch (IOException e) {
                    log.error("Error reading filter file {}: {}", path, e.getMessage());
                }
            });
        } catch (IOException e) {
            log.error("Error listing filters directory: {}", e.getMessage());
        }
        return foundModules;
    }

    private Set<String> syncRules() {
        Set<String> foundRules = new HashSet<>();
        Path rulesPath = Paths.get(".",Constants.APP_RULE_DEFINITIONS);
        if (!Files.exists(rulesPath) || !Files.isDirectory(rulesPath)) {
            log.warn("Rules directory not found: {}", Constants.APP_RULE_DEFINITIONS);
            return foundRules;
        }

        Yaml yaml = new Yaml();
        try (Stream<Path> paths = Files.walk(rulesPath)) {

            paths.filter(path -> Files.isRegularFile(path) && isYamlFile(path)).forEach(path -> {
                try {
                    String content = Files.readString(path);
                    RuleYaml ruleYaml = yaml.loadAs(content, RuleYaml.class);
                    if (ruleYaml == null || ruleYaml.getName() == null) {
                        log.warn("Skipping invalid rule file: {}", path);
                        return;
                    }

                    foundRules.add(ruleYaml.getName());
                    Optional<UtmCorrelationRules> ruleOpt = rulesRepository.findOneByRuleName(ruleYaml.getName());
                    UtmCorrelationRulesDTO ruleDto = new UtmCorrelationRulesDTO();

                    if (ruleOpt.isPresent()) {
                        ruleDto.setId(ruleOpt.get().getId());
                    } else {
                        ruleDto.setId(rulesRepository.getNextId());
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
                    ruleDto.setRuleActive(true);

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

                    UtmCorrelationRules entity = rulesMapper.toEntity(ruleDto);
                    if (ruleOpt.isPresent()) {
                        rulesService.updateRule(entity, true);

                    } else {
                        rulesService.save(entity, true);
                    }

                } catch (Exception e) {
                    log.error("Error processing rule file {}: {}", path, e.getMessage());
                }
            });
        } catch (IOException e) {
            log.error("Error walking rules directory: {}", e.getMessage());
        }
        return foundRules;
    }

    private void cleanupOrphanedFilters(Set<String> currentFilesystemModules) {
        if (currentFilesystemModules.isEmpty()) return;

        List<UtmLogstashFilter> systemFilters = filterRepository.findAllBySystemOwnerIsTrue();
        systemFilters.stream()
            .filter(filter -> !currentFilesystemModules.contains(filter.getModuleName()))
            .forEach(filter -> {
                log.info("Deleting orphaned system filter: {}", filter.getModuleName());
                filterService.delete(filter.getId());
            });
    }

    private void cleanupOrphanedRules(Set<String> currentFilesystemRules) {
        if (currentFilesystemRules.isEmpty()) return;

        List<UtmCorrelationRules> systemRules = rulesRepository.findAllBySystemOwnerIsTrue();
        systemRules.stream()
            .filter(rule -> !currentFilesystemRules.contains(rule.getRuleName()))
            .forEach(rule -> {
                log.info("Deleting orphaned system rule: {}", rule.getRuleName());
                try {
                    rulesService.deleteRule(rule.getId(), true);
                } catch (Exception e) {
                    log.error("Error deleting orphaned system rule {}: {}", rule.getRuleName(), e.getMessage());
                }
            });
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

    @Data
    public static class RuleYaml {
        private List<String> dataTypes;
        private String name;
        private ImpactYaml impact;
        private String category;
        private String technique;
        private AdversaryType adversary;
        private String description;
        private List<String> references;
        private String where;
        private List<com.park.utmstack.domain.correlation.rules.SearchRequest> afterEvents;
        private List<String> groupBy;
        private List<String> deduplicateBy;
    }

    @Data
    public static class ImpactYaml {
        private Integer confidentiality;
        private Integer integrity;
        private Integer availability;
    }
}
