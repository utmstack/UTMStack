package com.park.utmstack.repository;

import com.park.utmstack.domain.UtmServer;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.data.jpa.repository.Query;
import org.springframework.stereotype.Repository;


/**
 * Spring Data  repository for the UtmServer entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmServerRepository extends JpaRepository<UtmServer, Long>, JpaSpecificationExecutor<UtmServer> {
  @Query("select distinct min(us.id) as id from UtmServer us")
  Long getUtmServer();
}
