package com.park.utmstack.repository;

import com.park.utmstack.domain.UtmAlertTagRule;
import org.hibernate.jpa.TypedParameterValue;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;


/**
 * Spring Data  repository for the UtmTagRule entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmAlertTagRuleRepository extends JpaRepository<UtmAlertTagRule, Long> {

    @Query(nativeQuery = true, value = "select uatr.* from utm_alert_tag_rule uatr where " +
        "(:id is null or uatr.id = :id) " +
        "and (:name is null or lower(uatr.rule_name) like lower(concat('%', :name, '%'))) " +
        "and (:conditionField is null or lower(uatr.rule_conditions) like lower(concat('%', :conditionField, '%'))) " +
        "and (:conditionValue is null or lower(uatr.rule_conditions) like lower(concat('%', :conditionValue, '%'))) " +
        "and (:isActive is null or uatr.rule_active = :isActive) " +
        "and (:isDeleted is null or uatr.rule_deleted = :isDeleted) " +
        "and (:tagIds is null or (cast(string_to_array(cast(:tagIds as varchar) , ',') as int[]) && cast(string_to_array(uatr.rule_applied_tags , ',') as int[])))")
    Page<UtmAlertTagRule> findByFilter(@Param("id") TypedParameterValue id,
                                       @Param("name") TypedParameterValue name,
                                       @Param("conditionField") TypedParameterValue conditionField,
                                       @Param("conditionValue") TypedParameterValue conditionValue,
                                       @Param("tagIds") TypedParameterValue tagIds,
                                       @Param("isActive") TypedParameterValue isActive,
                                       @Param("isDeleted") TypedParameterValue isDeleted,
                                       Pageable pageable);

    List<UtmAlertTagRule> findAllByIdIn(List<Long> ids);
}
