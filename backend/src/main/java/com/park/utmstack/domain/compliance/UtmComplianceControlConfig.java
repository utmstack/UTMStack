package com.park.utmstack.domain.compliance;

import com.park.utmstack.domain.compliance.enums.ComplianceStrategy;
import lombok.Getter;
import lombok.Setter;
import org.hibernate.annotations.GenericGenerator;

import javax.persistence.*;
import javax.validation.constraints.Size;
import java.io.Serializable;
import java.util.ArrayList;
import java.util.List;

@Getter
@Setter
@Entity
@Table(name = "utm_compliance_control_config")
public class UtmComplianceControlConfig implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GenericGenerator(name = "CustomIdentityGenerator",
            strategy = "com.park.utmstack.util.CustomIdentityGenerator")
    @GeneratedValue(generator = "CustomIdentityGenerator")
    private Long id;

    @Column(name = "standard_section_id",
            nullable = false
    )
    private Long standardSectionId;

    @ManyToOne
    @JoinColumn(name = "standard_section_id",
            insertable = false,
            updatable = false
    )
    private UtmComplianceStandardSection section;

    @Column(name = "control_name",
            nullable = false
    )
    @Size(min = 10, max = 200)
    private String controlName;

    @Column(name = "control_solution",
            length = 2000
    )
    private String controlSolution;

    @Column(name = "control_remediation",
            length = 2000
    )
    private String controlRemediation;

    @Enumerated(EnumType.STRING)
    @Column(name = "control_strategy",
            nullable = false
    )
    private ComplianceStrategy controlStrategy;

    @OneToMany(
            mappedBy = "controlConfig",
            cascade = CascadeType.ALL,
            orphanRemoval = true
    )
    private List<UtmComplianceQueryConfig> queriesConfigs = new ArrayList<>();
}
