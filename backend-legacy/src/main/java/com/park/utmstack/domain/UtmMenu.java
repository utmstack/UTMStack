package com.park.utmstack.domain;

import com.fasterxml.jackson.databind.annotation.JsonDeserialize;
import com.fasterxml.jackson.databind.annotation.JsonSerialize;
import com.park.utmstack.domain.shared_types.MenuType;
import org.hibernate.annotations.GenericGenerator;

import javax.persistence.*;
import javax.validation.constraints.NotNull;
import javax.validation.constraints.Size;
import java.io.Serializable;
import java.util.List;

@Entity
@Table(name = "utm_menu")
public class UtmMenu implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GenericGenerator(name = "CustomIdentityGenerator", strategy = "com.park.utmstack.util.CustomIdentityGenerator")
    @GeneratedValue(generator = "CustomIdentityGenerator")
    private Long id;

    @NotNull
    @Size(min = 1, max = 50)
    @Column(name = "name", length = 50, unique = true, nullable = false)
    private String name;

    @Column(name = "url", length = 1024)
    private String url;

    @Column(name = "parent_id")
    private Long parentId;

    @Column(name = "type")
    private Integer type;

    @Column(name = "dashboard_id")
    private Long dashboardId;

    @Column(name = "position")
    private Long position;

    @Column(name = "menu_active")
    private Boolean menuActive;

    @Column(name = "menu_action")
    private Boolean menuAction;

    @Column(name = "menu_icon", length = 100)
    private String menuIcon;

    //    @Enumerated(EnumType.STRING)
    @Column(name = "module_name_short", length = 50)
    private String moduleNameShort;

    @Transient
    @JsonSerialize
    @JsonDeserialize
    private List<String> authorities;

    @ManyToOne
    @JoinColumn(name = "parent_id", referencedColumnName = "id", insertable = false, updatable = false)
    private UtmMenu parent;

    public UtmMenu() {
    }

    public UtmMenu(MenuType menu) {
        this.id = menu.getId();
        this.name = menu.getName();
        this.url = menu.getUrl();
        this.parentId = menu.getParentId();
        this.type = menu.getType();
        this.dashboardId = menu.getDashboardId();
        this.position = menu.getPosition();
        this.menuActive = menu.getMenuActive();
        this.menuAction = menu.getMenuAction();
        this.menuIcon = menu.getMenuIcon();
        this.moduleNameShort = menu.getModulesNameShort();
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

    public UtmMenu getParent() {
        return parent;
    }

    public void setParent(UtmMenu parent) {
        this.parent = parent;
    }

    public Integer getType() {
        return type;
    }

    public void setType(Integer type) {
        this.type = type;
    }

    public List<String> getAuthorities() {
        return authorities;
    }

    public void setAuthorities(List<String> authorities) {
        this.authorities = authorities;
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

    public Boolean getMenuActive() {
        return menuActive;
    }

    public void setMenuActive(Boolean menuActive) {
        this.menuActive = menuActive;
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

    public String getModuleNameShort() {
        return moduleNameShort;
    }

    public void setModuleNameShort(String moduleNameShort) {
        this.moduleNameShort = moduleNameShort;
    }
}
