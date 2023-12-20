package com.park.utmstack.service;

import com.park.utmstack.domain.UtmMenu;
import com.park.utmstack.domain.shared_types.MenuType;
import org.springframework.data.repository.query.Param;

import java.util.List;
import java.util.Optional;

public interface UtmMenuService {

    /**
     * Save a UtmMenu.
     *
     * @param menu the entity to save
     * @return the persisted entity
     */
    UtmMenu save(UtmMenu menu) throws Exception;

    List<UtmMenu> saveAll(List<UtmMenu> menus) throws Exception;

    /**
     * Get all UtmMenu.
     *
     * @return the list of entities
     */
    List<UtmMenu> findAll();

    /**
     * Get a UtmMenu by id
     *
     * @param id the id of the entity
     * @return the entity
     */
    Optional<UtmMenu> findById(Long id);

    /**
     * Delete a UtmMenu
     *
     * @param id the id of the entity
     */
    void delete(Long id);

    /**
     * @param parentId
     * @param authorities
     * @return
     */
    Optional<List<UtmMenu>> findMenusByAuthorities(Long parentId, List<String> authorities);

    List<MenuType> getMenus(boolean includeModulesMenus) throws Exception;

    Boolean saveMenuStructure(List<MenuType> menus) throws Exception;

    List<UtmMenu> findAllByModuleNameShort(String nameShort) throws Exception;

    void deleteSysMenusNotIn(@Param("ids") List<Long> ids) throws Exception;

}
