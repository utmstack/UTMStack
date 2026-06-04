package com.park.utmstack.domain.network_scan;


import javax.persistence.*;
import javax.validation.constraints.Size;
import java.io.Serializable;

/**
 * A UtmAssetTags.
 */
@Entity
@Table(name = "utm_asset_types")
public class UtmAssetTypes implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Size(max = 100)
    @Column(name = "type_name", length = 100, unique = true)
    private String typeName;

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getTypeName() {
        return typeName;
    }

    public void setTypeName(String typeName) {
        this.typeName = typeName;
    }
}
