package com.park.utmstack.service.compliance.config;

import com.park.utmstack.domain.compliance.UtmComplianceStandardSection;
import com.park.utmstack.repository.compliance.UtmComplianceStandardSectionRepository;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;
import java.util.Optional;

@Service
public class UtmComplianceStandardSectionService {

    private final UtmComplianceStandardSectionRepository standardSectionRepository;

    public UtmComplianceStandardSectionService(UtmComplianceStandardSectionRepository standardSectionRepository) {
        this.standardSectionRepository = standardSectionRepository;
    }

    public UtmComplianceStandardSection save(UtmComplianceStandardSection complianceStandardSection) {
        return standardSectionRepository.save(complianceStandardSection);
    }

    public void saveAll(List<UtmComplianceStandardSection> sections) {
        standardSectionRepository.saveAll(sections);
    }

    public UtmComplianceStandardSection saveAndFlush(UtmComplianceStandardSection complianceStandardSection) {
        return standardSectionRepository.saveAndFlush(complianceStandardSection);
    }

    @Transactional(readOnly = true)
    public Page<UtmComplianceStandardSection> findAll(Pageable pageable) {
        return standardSectionRepository.findAll(pageable);
    }

    @Transactional(readOnly = true)
    public Optional<UtmComplianceStandardSection> findOne(Long id) {
        return standardSectionRepository.findById(id);
    }

    public void delete(Long id) {
        standardSectionRepository.deleteById(id);
    }

    public List<UtmComplianceStandardSection> getSectionsWithReportsByStandard(Long standardId) {
        return standardSectionRepository.getSectionsWithReportsByStandard(standardId);
    }

    public Optional<UtmComplianceStandardSection> findByStandardSectionNameLike(String sectionName) {
        return standardSectionRepository.findByStandardSectionNameLike(sectionName);
    }

    public List<UtmComplianceStandardSection> findAllByStandardSectionNameNotIn(List<String> sectionNames) {
        return standardSectionRepository.findAllByStandardSectionNameNotIn(sectionNames);
    }

    public List<UtmComplianceStandardSection> findAllByStandardIdAndStandardSectionNameNotIn(Long standardId, List<String> names) {
        return standardSectionRepository.findAllByStandardIdAndStandardSectionNameNotIn(standardId, names);
    }

    @Transactional
    public void deleteSectionsByStandardAndIdNotIn(Long standardId, List<Long> sectionIds) {
        standardSectionRepository.deleteSectionsByStandardAndIdNotIn(standardId, sectionIds);
    }

    public Long getSystemSequenceNextValue(Long standardId) {
        long value = standardId + 1;
        Optional<UtmComplianceStandardSection> opt = standardSectionRepository.findFirstByStandardIdOrderByIdDesc(standardId);
        if (opt.isPresent())
            value = opt.get().getId() + 1;
        return value;
    }
}
