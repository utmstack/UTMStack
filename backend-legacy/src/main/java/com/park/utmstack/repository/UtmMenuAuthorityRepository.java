package com.park.utmstack.repository;

import com.park.utmstack.domain.UtmMenuAuthority;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Modifying;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;


/**
 * Spring Data  repository for the UtmMenuAuthority entity.
 */
@SuppressWarnings("unused")
@Repository
public interface UtmMenuAuthorityRepository extends JpaRepository<UtmMenuAuthority, Long> {

    void deleteAllByMenuIdEquals(Long menuId);

    void deleteAllByMenuIdIn(List<Long> menuIds);

    void deleteAllByIdIn(List<Long> menuIds);

    Optional<List<UtmMenuAuthority>> findAllByMenuIdEquals(Long menuId);

    @Modifying
    @Query("delete from UtmMenuAuthority a where a.menuId = :menuId and a.authorityName not in :authorities")
    void deleteMenuAuthoritiesNotIn(@Param("menuId") Long menuId, @Param("authorities") List<String> authorities);
}
