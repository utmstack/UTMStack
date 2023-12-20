package com.park.utmstack.repository;

import com.park.utmstack.domain.UtmAlertTag;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.stereotype.Repository;

import java.util.List;


/**
 * Spring Data  repository for the UtmAlertCategory entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmAlertTagRepository extends JpaRepository<UtmAlertTag, Long>, JpaSpecificationExecutor<UtmAlertTag> {
    List<UtmAlertTag> findAllByIdIn(List<Long> ids);
}
