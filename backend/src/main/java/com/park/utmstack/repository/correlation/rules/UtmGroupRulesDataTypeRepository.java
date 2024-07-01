package com.park.utmstack.repository.correlation.rules;

import com.park.utmstack.domain.correlation.rules.UtmGroupRulesDataType;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.stereotype.Repository;

import java.util.List;

/**
 * Spring Data  repository for the UtmGroupRulesDataType entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmGroupRulesDataTypeRepository extends JpaRepository<UtmGroupRulesDataType, Long> {
    @Query(nativeQuery = true, value = "SELECT nextval('utm_data_types_id_seq')")
    Long getNextId();

    List<UtmGroupRulesDataType> findByRuleId(Long ruleId);
}
