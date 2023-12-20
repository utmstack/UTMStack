package com.park.utmstack.repository.logstash_filter;

import com.park.utmstack.domain.logstash_filter.UtmLogstashFilterGroup;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.stereotype.Repository;

import java.util.List;


/**
 * Spring Data  repository for the UtmLogstashFilterGroup entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmLogstashFilterGroupRepository extends JpaRepository<UtmLogstashFilterGroup, Long>, JpaSpecificationExecutor<UtmLogstashFilterGroup> {

    void deleteAllBySystemOwnerIsTrueAndIdNotIn(List<Long> ids);


}
