package com.park.utmstack.domain;


import javax.persistence.*;
import javax.validation.constraints.NotNull;
import java.io.Serializable;

/**
 * A UtmMenuAuthority.
 */
@Entity
@Table(name = "utm_menu_authority")
public class UtmMenuAuthority implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @NotNull
    @Column(name = "menu_id", nullable = false)
    private Long menuId;

    @NotNull
    @Column(name = "authority_name", nullable = false)
    private String authorityName;

    @ManyToOne
    @JoinColumn(name = "menu_id", referencedColumnName = "id", insertable = false, updatable = false)
    private UtmMenu menu;

    public UtmMenuAuthority() {
    }

    public UtmMenuAuthority(Long menuId, String authorityName) {
        this.menuId = menuId;
        this.authorityName = authorityName;
    }

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public Long getMenuId() {
        return menuId;
    }

    public UtmMenuAuthority menuId(Long menuId) {
        this.menuId = menuId;
        return this;
    }

    public void setMenuId(Long menuId) {
        this.menuId = menuId;
    }

    public String getAuthorityName() {
        return authorityName;
    }

    public UtmMenuAuthority authorityName(String authorityName) {
        this.authorityName = authorityName;
        return this;
    }

    public void setAuthorityName(String authorityName) {
        this.authorityName = authorityName;
    }

    public UtmMenu getMenu() {
        return menu;
    }

    public void setMenu(UtmMenu menu) {
        this.menu = menu;
    }
}
