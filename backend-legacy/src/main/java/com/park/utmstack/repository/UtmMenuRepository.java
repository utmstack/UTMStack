package com.park.utmstack.repository;

import com.park.utmstack.domain.UtmMenu;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.data.jpa.repository.Modifying;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;

@Repository
public interface UtmMenuRepository extends JpaRepository<UtmMenu, Long>, JpaSpecificationExecutor<UtmMenu> {

    @Query(nativeQuery = true, value = "SELECT * FROM utm_menu WHERE utm_menu.parent_id = :parentId AND utm_menu.id IN (SELECT utm_menu_authority.menu_id FROM utm_menu_authority WHERE utm_menu_authority.authority_name IN (:authorities))")
    Optional<List<UtmMenu>> findMenusByAuthorities(@Param("parentId") Long parentId,
                                                   @Param("authorities") List<String> authorities);

    List<UtmMenu> findAllByParentIdIsNull();

    List<UtmMenu> findAllByParentId(long parent);

    @Query(nativeQuery = true, value = "select utm_menu.* from utm_menu where :nameShort = any(string_to_array(utm_menu.module_name_short, ','))")
    List<UtmMenu> findAllByModuleNameShort(@Param("nameShort") String nameShort);

    @Modifying
    @Query("delete from UtmMenu m where m.type = 1 and m.id not in :ids")
    void deleteSysMenusNotIn(@Param("ids") List<Long> ids);
}
