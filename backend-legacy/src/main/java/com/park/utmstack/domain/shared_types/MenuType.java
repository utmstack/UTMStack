package com.park.utmstack.domain.shared_types;

import com.park.utmstack.domain.UtmMenu;
import org.springframework.util.CollectionUtils;

import java.util.ArrayList;
import java.util.List;

public class MenuType {
    private Long id;
    private String name;
    private String url;
    private Long parentId;
    private Integer type;
    private Long dashboardId;
    private Long position;
    private Boolean menuActive;
    private Boolean menuAction;
    private String menuIcon;
    private String modulesNameShort;
    private List<String> authorities;
    private List<MenuType> childrens;
    private List<MenuType> actions;

    public MenuType() {
    }

    public MenuType(UtmMenu menu) {
        this.id = menu.getId();
        this.name = menu.getName();
        this.url = menu.getUrl();
        this.parentId = menu.getParentId();
        this.type = menu.getType();
        this.dashboardId = menu.getDashboardId();
        this.position = menu.getPosition();
        this.menuActive = menu.getMenuActive();
        this.menuAction = menu.getMenuAction();
        this.authorities = menu.getAuthorities();
        this.menuIcon = menu.getMenuIcon();
        this.modulesNameShort = menu.getModuleNameShort();
    }

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getUrl() {
        return url;
    }

    public void setUrl(String url) {
        this.url = url;
    }

    public Long getParentId() {
        return parentId;
    }

    public void setParentId(Long parentId) {
        this.parentId = parentId;
    }

    public Integer getType() {
        return type;
    }

    public void setType(Integer type) {
        this.type = type;
    }

    public Long getDashboardId() {
        return dashboardId;
    }

    public void setDashboardId(Long dashboardId) {
        this.dashboardId = dashboardId;
    }

    public Long getPosition() {
        return position;
    }

    public void setPosition(Long position) {
        this.position = position;
    }

    public List<String> getAuthorities() {
        return authorities;
    }

    public void setAuthorities(List<String> authorities) {
        this.authorities = authorities;
    }

    public List<MenuType> getChildrens() {
        return childrens;
    }

    public void setChildrens(List<MenuType> children) {
        this.childrens = children;
    }

    public Boolean getMenuActive() {
        return menuActive;
    }

    public void setMenuActive(Boolean menuActive) {
        this.menuActive = menuActive;
    }

    public void addChildren(MenuType child) {
        if (CollectionUtils.isEmpty(childrens))
            childrens = new ArrayList<>();
        childrens.add(child);
    }

    public List<MenuType> getActions() {
        return actions;
    }

    public void setActions(List<MenuType> actions) {
        this.actions = actions;
    }

    public void addAction(MenuType action) {
        if (CollectionUtils.isEmpty(actions))
            actions = new ArrayList<>();
        actions.add(action);
    }

    public Boolean getMenuAction() {
        return menuAction;
    }

    public void setMenuAction(Boolean menuAction) {
        this.menuAction = menuAction;
    }

    public String getMenuIcon() {
        return menuIcon;
    }

    public void setMenuIcon(String menuIcon) {
        this.menuIcon = menuIcon;
    }

    public String getModulesNameShort() {
        return modulesNameShort;
    }

    public void setModulesNameShort(String modulesNameShort) {
        this.modulesNameShort = modulesNameShort;
    }
}
