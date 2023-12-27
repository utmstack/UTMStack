package com.park.utmstack.service.impl;

import com.park.utmstack.domain.UtmMenu;
import com.park.utmstack.domain.UtmMenuAuthority;
import com.park.utmstack.domain.shared_types.MenuType;
import com.park.utmstack.repository.UtmMenuRepository;
import com.park.utmstack.service.UtmMenuAuthorityService;
import com.park.utmstack.service.UtmMenuService;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.CollectionUtils;

import javax.persistence.EntityManager;
import java.util.*;
import java.util.stream.Collectors;

@Service
@Transactional
public class UtmMenuServiceImpl implements UtmMenuService {

    private static final String CLASS_NAME = "UtmMenuServiceImpl";

    private final UtmMenuRepository menuRepository;
    private final UtmMenuAuthorityService menuAuthorityService;
    private final EntityManager em;

    public UtmMenuServiceImpl(UtmMenuRepository menuRepository, UtmMenuAuthorityService menuAuthorityService, EntityManager em) {
        this.menuRepository = menuRepository;
        this.menuAuthorityService = menuAuthorityService;
        this.em = em;
    }

    @Override
    public UtmMenu save(UtmMenu menu) throws Exception {
        final String ctx = CLASS_NAME + ".save";

        try {
            if (menu.getPosition() == null) {
                StringBuilder query = new StringBuilder("SELECT MAX(utm_menu.position) FROM utm_menu WHERE parent_id ");

                if (menu.getParentId() == null)
                    query.append("IS NULL");
                else
                    query.append("= ").append(menu.getParentId());

                Short lastPosition = (Short) em.createNativeQuery(query.toString()).getSingleResult();
                menu.setPosition(lastPosition == null ? 0 : lastPosition.longValue() + 1);
            }

            if (menu.getMenuAction() == null)
                menu.setMenuAction(false);

            return menuRepository.save(menu);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    @Override
    public List<UtmMenu> saveAll(List<UtmMenu> menus) throws Exception {
        final String ctx = CLASS_NAME + ".saveAll";
        try {
            return menuRepository.saveAll(menus);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    @Override
    @Transactional(readOnly = true)
    public List<UtmMenu> findAll() {
        return menuRepository.findAll();
    }

    @Override
    @Transactional(readOnly = true)
    public Optional<UtmMenu> findById(Long id) {
        return menuRepository.findById(id);
    }

    @Override
    public void delete(Long id) {
        menuRepository.deleteById(id);
    }

    @Override
    @Transactional
    public Optional<List<UtmMenu>> findMenusByAuthorities(Long parentId, List<String> authorities) {
        return menuRepository.findMenusByAuthorities(parentId, authorities);
    }

    /**
     * @return
     * @throws Exception
     */
    public List<MenuType> getMenus(boolean includeModulesMenus) throws Exception {
        final String ctx = CLASS_NAME + ".getMenus";
        try {
            List<UtmMenu> parents = menuRepository.findAllByParentIdIsNull();

            if (!includeModulesMenus)
                parents = parents.stream().filter(menu -> Objects.isNull(menu.getModuleNameShort()))
                    .collect(Collectors.toList());

            if (CollectionUtils.isEmpty(parents))
                return Collections.emptyList();

            List<MenuType> menus = new ArrayList<>();

            for (UtmMenu parent : parents) {
                MenuType menu = new MenuType(parent);
                menu.setAuthorities(buildAuthorities(parent.getId()));

                List<UtmMenu> children = menuRepository.findAllByParentId(parent.getId());

                if (CollectionUtils.isEmpty(children)) {
                    menus.add(menu);
                    continue;
                }

                children.forEach(child -> {
                    if (!child.getMenuAction()) {
                        MenuType childMenu = new MenuType(child);
                        childMenu.setAuthorities(buildAuthorities(child.getId()));
                        menu.addChildren(childMenu);
                    }
                });

                List<UtmMenu> actions = children.stream().filter(UtmMenu::getMenuAction).collect(Collectors.toList());

                if (!CollectionUtils.isEmpty(actions)) {
                    actions.forEach(child -> menu.addAction(new MenuType(child)));
                    menu.getActions().sort(Comparator.comparing(MenuType::getPosition));
                }

                if (!CollectionUtils.isEmpty(menu.getChildrens()))
                    menu.getChildrens().sort(Comparator.comparing(MenuType::getPosition));
                menus.add(menu);
            }
            menus.sort(Comparator.comparing(MenuType::getPosition));
            return menus;
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    private List<String> buildAuthorities(Long menuId) {
        Optional<List<UtmMenuAuthority>> authorities = menuAuthorityService.findAllByMenuIdEquals(menuId);
        List<String> result = new ArrayList<>();
        authorities.ifPresent(utmMenuAuthorities -> {
            utmMenuAuthorities.forEach(utmMenuAuthority -> result.add(utmMenuAuthority.getAuthorityName()));
        });
        return result;
    }

    /**
     * @param menus
     */
    public Boolean saveMenuStructure(List<MenuType> menus) throws Exception {
        final String ctx = CLASS_NAME + ".saveMenuStructure";

        try {
            if (CollectionUtils.isEmpty(menus))
                return true;

            for (MenuType menu : menus) {
                UtmMenu dbMenu = updateMenuStructure(menu);

                if (Objects.isNull(menu.getParentId()))
                    continue;

                List<MenuType> subMenus = menu.getChildrens();

                for (MenuType subMenu : subMenus) {
                    subMenu.setParentId(dbMenu.getId());
                    updateMenuStructure(subMenu);
                }
            }
            return true;
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    private UtmMenu updateMenuStructure(MenuType menu) throws Exception {
        final String ctx = CLASS_NAME + ".updateMenuStructure";
        try {
            return save(new UtmMenu(menu));
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    @Override
    public List<UtmMenu> findAllByModuleNameShort(String nameShort) throws Exception {
        final String ctx = CLASS_NAME + ".findAllByModuleNameShort";
        try {
            return menuRepository.findAllByModuleNameShort(nameShort);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    @Override
    public void deleteSysMenusNotIn(List<Long> ids) throws Exception {
        final String ctx = CLASS_NAME + ".findAllByModuleNameShort";
        try {
            menuRepository.deleteSysMenusNotIn(ids);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }

    }
}
