package com.park.utmstack.service.compliance.config;

import com.park.utmstack.domain.compliance.UtmComplianceStandard;
import com.park.utmstack.repository.compliance.UtmComplianceStandardRepository;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;
import java.util.Optional;

@Service
public class UtmComplianceStandardService {

    private final UtmComplianceStandardRepository complianceStandardRepository;

    public UtmComplianceStandardService(UtmComplianceStandardRepository complianceStandardRepository) {
        this.complianceStandardRepository = complianceStandardRepository;
    }

    public UtmComplianceStandard save(UtmComplianceStandard complianceStandard) {
        return complianceStandardRepository.save(complianceStandard);
    }

    public UtmComplianceStandard saveAndFlush(UtmComplianceStandard complianceStandard) {
        return complianceStandardRepository.saveAndFlush(complianceStandard);
    }

    @Transactional
    public void deleteAllBySystemOwnerIsTrueAndIdNotIn(List<Long> standardIds) {
        complianceStandardRepository.deleteAllBySystemOwnerIsTrueAndIdNotIn(standardIds);
    }

    @Transactional(readOnly = true)
    public Page<UtmComplianceStandard> findAll(Pageable pageable) {
        return complianceStandardRepository.findAll(pageable);
    }

    @Transactional(readOnly = true)
    public Optional<UtmComplianceStandard> findOne(Long id) {
        return complianceStandardRepository.findById(id);
    }

    public void delete(Long id) {
        complianceStandardRepository.deleteById(id);
    }

    @Transactional(readOnly = true)
    public Optional<UtmComplianceStandard> findByStandardNameLike(String standardName) {
        return complianceStandardRepository.findByStandardNameLike(standardName);
    }

    public Long getSystemSequenceNextValue() {
        long value = 100;
        Optional<UtmComplianceStandard> opt = complianceStandardRepository.findFirstBySystemOwnerIsTrueOrderByIdDesc();
        if (opt.isPresent())
            value = opt.get().getId() + 100;
        return value;
    }
}
