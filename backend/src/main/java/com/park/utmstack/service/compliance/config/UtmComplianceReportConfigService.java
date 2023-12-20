package com.park.utmstack.service.compliance.config;

import com.park.utmstack.domain.compliance.UtmComplianceReportConfig;
import com.park.utmstack.domain.compliance.UtmComplianceStandard;
import com.park.utmstack.domain.compliance.UtmComplianceStandardSection;
import com.park.utmstack.repository.compliance.UtmComplianceReportConfigRepository;
import com.park.utmstack.service.chart_builder.UtmDashboardService;
import com.park.utmstack.util.UtilPagination;
import com.park.utmstack.util.exceptions.UtmPageNumberNotSupported;
import org.springframework.dao.DataIntegrityViolationException;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.CollectionUtils;
import org.springframework.util.StringUtils;

import javax.persistence.EntityManager;
import java.util.List;
import java.util.Objects;
import java.util.Optional;

@Service
public class UtmComplianceReportConfigService {

    private static final String CLASSNAME = "UtmComplianceReportConfigService";
    private final UtmComplianceReportConfigRepository complianceReportConfigRepository;
    private final UtmDashboardService dashboardService;
    private final UtmComplianceStandardService standardService;
    private final UtmComplianceStandardSectionService standardSectionService;
    private final EntityManager em;

    public UtmComplianceReportConfigService(UtmComplianceReportConfigRepository complianceReportConfigRepository,
                                            UtmDashboardService dashboardService,
                                            UtmComplianceStandardService standardService,
                                            UtmComplianceStandardSectionService standardSectionService,
                                            EntityManager em) {
        this.complianceReportConfigRepository = complianceReportConfigRepository;
        this.dashboardService = dashboardService;
        this.standardService = standardService;
        this.standardSectionService = standardSectionService;
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

    public List<UtmComplianceReportConfig> getReportsByFilters(Long standardId, String solution, Long sectionId, Pageable pageable) throws
        UtmPageNumberNotSupported {
        StringBuilder script = new StringBuilder(
            "SELECT cfg.* FROM utm_compliance_report_config cfg INNER JOIN utm_compliance_standard_section sec ON cfg.standard_section_id=sec.id INNER JOIN utm_compliance_standard st ON sec.standard_id=st.id");

        boolean hasWhere = false;

        if (standardId != null) {
            hasWhere = true;
            script.append(" WHERE").append(" st.id = ").append(standardId);
        }

        if (StringUtils.hasText(solution)) {
            String condition = "cfg.config_solution ILIKE '%" + solution + "%'";
            script.append(hasWhere ? " AND " : " WHERE ").append(condition);
            hasWhere = true;
        }

        if (sectionId != null) {
            String condition = "sec.id = " + sectionId;
            script.append(hasWhere ? " AND " : " WHERE ").append(condition);
        }

        return em.createNativeQuery(script.toString(), UtmComplianceReportConfig.class).setFirstResult(
            UtilPagination.getFirstForNativeSql(pageable.getPageSize(), pageable.getPageNumber())).setMaxResults(
            pageable.getPageSize()).getResultList();
    }

    public void deleteAllByConfigSolutionAndSectionIdAndDashboardId(String configSolution, Long sectionId, Long dashboardId) {
        complianceReportConfigRepository.deleteAllByConfigSolutionAndStandardSectionIdAndDashboardId(configSolution, sectionId, dashboardId);
    }

    /**
     * @param reports
     * @param override
     */
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
}
