package com.park.utmstack.repository.network_scan;

import com.park.utmstack.domain.network_scan.UtmAssetTypes;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;


/**
 * Spring Data  repository for the UtmAssetTags entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmAssetTypesRepository extends JpaRepository<UtmAssetTypes, Long> {

}
