package com.park.utmstack.repository;

import com.park.utmstack.domain.UtmImages;
import com.park.utmstack.domain.shared_types.enums.ImageShortName;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Modifying;
import org.springframework.data.jpa.repository.Query;
import org.springframework.stereotype.Repository;


/**
 * Spring Data  repository for the UtmImages entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmImagesRepository extends JpaRepository<UtmImages, ImageShortName> {

    @Modifying
    @Query("update UtmImages i set i.userImg = i.systemImg")
    void reset();

}
