package com.park.utmstack.domain.chart_builder;


import javax.persistence.*;
import javax.validation.constraints.NotNull;
import javax.validation.constraints.Size;
import java.io.Serializable;
import java.util.Objects;

/**
 * A UtmDashboardAuthority.
 */
@Entity
@Table(name = "utm_dashboard_authority")
public class UtmDashboardAuthority implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @NotNull
    @Column(name = "id_dashboard", nullable = false)
    private Long idDashboard;

    @NotNull
    @Size(max = 50)
    @Column(name = "authority_name", length = 50, nullable = false)
    private String authorityName;

    // jhipster-needle-entity-add-field - JHipster will add fields here, do not remove
    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public Long getIdDashboard() {
        return idDashboard;
    }

    public UtmDashboardAuthority idDashboard(Long idDashboard) {
        this.idDashboard = idDashboard;
        return this;
    }

    public void setIdDashboard(Long idDashboard) {
        this.idDashboard = idDashboard;
    }

    public String getAuthorityName() {
        return authorityName;
    }

    public UtmDashboardAuthority authorityName(String authorityName) {
        this.authorityName = authorityName;
        return this;
    }

    public void setAuthorityName(String authorityName) {
        this.authorityName = authorityName;
    }
    // jhipster-needle-entity-add-getters-setters - JHipster will add getters and setters here, do not remove

    @Override
    public boolean equals(Object o) {
        if (this == o) {
            return true;
        }
        if (o == null || getClass() != o.getClass()) {
            return false;
        }
        UtmDashboardAuthority utmDashboardAuthority = (UtmDashboardAuthority) o;
        if (utmDashboardAuthority.getId() == null || getId() == null) {
            return false;
        }
        return Objects.equals(getId(), utmDashboardAuthority.getId());
    }

    @Override
    public int hashCode() {
        return Objects.hashCode(getId());
    }

    @Override
    public String toString() {
        return "UtmDashboardAuthority{" +
            "id=" + getId() +
            ", idDashboard=" + getIdDashboard() +
            ", authorityName='" + getAuthorityName() + "'" +
            "}";
    }
}
