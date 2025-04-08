package com.park.utmstack.service.compliance.config;

import com.fasterxml.jackson.databind.node.ObjectNode;
import com.park.utmstack.domain.chart_builder.UtmDashboard;
import com.park.utmstack.domain.chart_builder.UtmDashboardVisualization;
import com.park.utmstack.domain.chart_builder.UtmVisualization;
import com.park.utmstack.domain.chart_builder.types.ChartType;
import com.park.utmstack.domain.compliance.UtmComplianceReportConfig;
import com.park.utmstack.domain.compliance.UtmComplianceStandard;
import com.park.utmstack.domain.compliance.UtmComplianceStandardSection;
import com.park.utmstack.domain.compliance.enums.ComplianceStatus;
import com.park.utmstack.domain.index_pattern.UtmIndexPattern;
import com.park.utmstack.repository.compliance.UtmComplianceReportConfigRepository;
import com.park.utmstack.service.chart_builder.UtmDashboardService;
import com.park.utmstack.service.chart_builder.UtmDashboardVisualizationService;
import com.park.utmstack.service.elasticsearch.ElasticsearchService;
import com.park.utmstack.service.elasticsearch.SearchUtil;
import com.park.utmstack.service.index_pattern.UtmIndexPatternService;
import com.park.utmstack.util.UtilPagination;
import com.park.utmstack.util.chart_builder.elasticsearch_dsl.requests.RequestDsl;
import com.park.utmstack.util.exceptions.UtmElasticsearchException;
import org.opensearch.client.opensearch.core.SearchResponse;
import org.springframework.dao.DataIntegrityViolationException;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.PageImpl;
import org.springframework.data.domain.PageRequest;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.CollectionUtils;
import org.springframework.util.StringUtils;

import javax.persistence.EntityManager;
import java.util.Collections;
import java.util.List;
import java.util.Objects;
import java.util.Optional;
import java.util.stream.Collectors;

@Service
public class UtmComplianceReportConfigService {

    private static final String CLASSNAME = "UtmComplianceReportConfigService";
    private final UtmComplianceReportConfigRepository complianceReportConfigRepository;
    private final UtmDashboardService dashboardService;
    private final UtmComplianceStandardService standardService;
    private final UtmComplianceStandardSectionService standardSectionService;

    private final UtmDashboardVisualizationService dashboardVisualizationService;

    private final ElasticsearchService elasticsearchService;
    private final EntityManager em;

    public UtmComplianceReportConfigService(UtmComplianceReportConfigRepository complianceReportConfigRepository,
                                            UtmDashboardService dashboardService,
                                            UtmComplianceStandardService standardService,
                                            UtmComplianceStandardSectionService standardSectionService,
                                            UtmDashboardVisualizationService dashboardVisualizationService, ElasticsearchService elasticsearchService,
                                            EntityManager em, UtmIndexPatternService indexPatternService) {
        this.complianceReportConfigRepository = complianceReportConfigRepository;
        this.dashboardService = dashboardService;
        this.standardService = standardService;
        this.standardSectionService = standardSectionService;
        this.dashboardVisualizationService = dashboardVisualizationService;
        this.elasticsearchService = elasticsearchService;
        this.em = em;
    }

    public UtmComplianceReportConfig save(UtmComplianceReportConfig complianceReportConfig) {
        return complianceReportConfigRepository.save(complianceReportConfig);
    }

    public void saveAll(List<UtmComplianceReportConfig> reports) {
        complianceReportConfigRepository.saveAll(reports);
    }

    @Transactional(readOnly = true)
    public Page<UtmComplianceReportConfig> findAll(Pageable pageable) {
        return complianceReportConfigRepository.findAll(pageable);
    }

    @Transactional(readOnly = true)
    public Optional<UtmComplianceReportConfig> findOne(Long id) {
        return complianceReportConfigRepository.findById(id);
    }

    @Transactional(readOnly = true)
    public Optional<UtmComplianceReportConfig> findByConfigSolutionAndStandardSectionIdAndDashboardId(String configSolution, Long sectionId, Long dashboardId) {
        return complianceReportConfigRepository.findByConfigSolutionAndStandardSectionIdAndDashboardId(configSolution, sectionId, dashboardId);
    }

    /**
     * Delete the logAnalyzerQuery by id.
     *
     * @param id the id of the entity
     */
    public void delete(Long id) {
        complianceReportConfigRepository.deleteById(id);
    }

    @Transactional
    public void deleteReportsByStandardIdAndIdNotIn(Long standardId, List<Long> reportIds) {
        complianceReportConfigRepository.deleteReportsByStandardIdAndIdNotIn(standardId, reportIds);
    }

    public Page<UtmComplianceReportConfig> getReportsByFilters(Long standardId, String solution, Long sectionId,
                                                               String search, Boolean expandDashboard, Boolean setStatus, Pageable pageable) throws UtmElasticsearchException {

        Page<UtmComplianceReportConfig> page = complianceReportConfigRepository.getReportsByFilters(standardId, solution, sectionId, search, pageable);

        if (expandDashboard != null && expandDashboard) {
            for (UtmComplianceReportConfig report : page) {
                dashboardVisualizationService.findAllByIdDashboard(report.getDashboardId()).ifPresent(report::setDashboard);
            }
        }

        if (setStatus) {
            for (UtmComplianceReportConfig report : page) {
                report.setConfigReportStatus(this.getStatus(report.getAssociatedDashboard()));
            }
        }

        return page;

    }


    public void deleteAllByConfigSolutionAndSectionIdAndDashboardId(String configSolution, Long sectionId, Long dashboardId) {
        complianceReportConfigRepository.deleteAllByConfigSolutionAndStandardSectionIdAndDashboardId(configSolution, sectionId, dashboardId);
    }

    public void importReports(List<UtmComplianceReportConfig> reports, boolean override) throws Exception {
        final String ctx = CLASSNAME + ".importReports";
        try {
            if (CollectionUtils.isEmpty(reports))
                return;

            for (UtmComplianceReportConfig report : reports) {
                Objects.requireNonNull(report.getSection(), String.format("Missing standard section for report: %1$s", report.getConfigSolution()));
                Objects.requireNonNull(report.getSection().getStandard(), String.format("Missing standard for section: %1$s", report.getSection().getStandardSectionName()));

                UtmComplianceStandardSection section = report.getSection();
                UtmComplianceStandard standard = report.getSection().getStandard();

                if (!Objects.isNull(report.getDashboard()))
                    dashboardService.importDashboards(report.getDashboard(), override);

                report.setId(null);

                // Standards
                Optional<UtmComplianceStandard> eStandard = standardService.findByStandardNameLike(
                        standard.getStandardName());

                if (eStandard.isPresent())
                    standard.setId(eStandard.get().getId());
                standard = standardService.saveAndFlush(standard);

                // Sections
                Optional<UtmComplianceStandardSection> eSection = standardSectionService.findByStandardSectionNameLike(
                        section.getStandardSectionName());

                if (eSection.isPresent())
                    section.setId(eSection.get().getId());
                else
                    section.setStandardId(standard.getId());

                section = standardSectionService.saveAndFlush(section);

                // Compliance report
                Optional<UtmComplianceReportConfig> eReport = complianceReportConfigRepository.
                        findByConfigSolutionAndStandardSectionIdAndDashboardId(report.getConfigSolution(), report.getStandardSectionId(), report.getDashboardId());

                eReport.ifPresent(utmComplianceReportConfig -> report.setId(utmComplianceReportConfig.getId()));
                report.setStandardSectionId(section.getId());
                save(report);
            }
        } catch (DataIntegrityViolationException e) {
            String msg = ctx + ": " + e.getMostSpecificCause().getMessage().replaceAll("\n", "");
            throw new Exception(msg);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            throw new Exception(msg);
        }
    }

    private ComplianceStatus getStatus(UtmDashboard dashboard) throws UtmElasticsearchException {
        List<UtmDashboardVisualization> dashboardVisualizations = dashboardVisualizationService.findAllByIdDashboard(dashboard.getId())
                .orElse(Collections.emptyList());

        UtmVisualization visualization = dashboardVisualizations.stream().filter(d -> d.getVisualization().getChartType().equals(ChartType.LIST_CHART)
                        || d.getVisualization().getChartType().equals(ChartType.TABLE_CHART))
                .map(UtmDashboardVisualization::getVisualization)
                .findFirst()
                .orElse(null);

        if(Objects.nonNull(visualization)){
            RequestDsl requestQuery = new RequestDsl(visualization);
            SearchResponse<ObjectNode> result = elasticsearchService.search(requestQuery.getSearchSourceBuilderForCount().build(), ObjectNode.class);
            return result.hits().total().value() > 0 ? ComplianceStatus.COMPLAINT : ComplianceStatus.NON_COMPLAINT;
        } else {
            return ComplianceStatus.NON_COMPLAINT;
        }
    }
}
