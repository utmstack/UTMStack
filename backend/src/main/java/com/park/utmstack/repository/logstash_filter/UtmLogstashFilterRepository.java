package com.park.utmstack.repository.logstash_filter;

import com.park.utmstack.domain.logstash_filter.UtmLogstashFilter;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;


/**
 * Spring Data  repository for the UtmLogstashFilter entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmLogstashFilterRepository extends JpaRepository<UtmLogstashFilter, Long>, JpaSpecificationExecutor<UtmLogstashFilter> {

    void deleteAllBySystemOwnerIsTrueAndIdNotIn(List<Long> ids);

    @Query(nativeQuery = true, value = "select utm_logstash_filter.* from utm_logstash_filter where :nameShort = any(string_to_array(utm_logstash_filter.module_name, ','))")
    List<UtmLogstashFilter> findAllByModuleName(@Param("nameShort") String nameShort);

    @Query("select ulf from UtmLogstashFilter ulf where ulf.id in (:filterList) and ulf.systemOwner=false")
    List<UtmLogstashFilter> findAllByListOfId(@Param("filterList") List<Long> filterList);

    @Query(nativeQuery = true, value = "select ulf.* from (select distinct filters.filter_id " +
        "from utm_group_logstash_pipeline_filters filters where filters.pipeline_id = :pipelineId) bypipeline " +
        "inner join utm_logstash_filter ulf on bypipeline.filter_id = ulf.id")
    List<UtmLogstashFilter> filtersByPipelineId(@Param("pipelineId") Long pipelineId);
}
