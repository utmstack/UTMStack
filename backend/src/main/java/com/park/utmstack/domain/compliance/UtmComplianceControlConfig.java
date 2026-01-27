package com.park.utmstack.domain.compliance;

import com.park.utmstack.domain.compliance.enums.ComplianceStrategy;
import lombok.Getter;
import lombok.Setter;
import org.hibernate.annotations.GenericGenerator;

import javax.persistence.*;
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

    @Column(name = "standard_section_id")
    private Long standardSectionId;

    @ManyToOne
    @JoinColumn(name = "standard_section_id",
            insertable = false,
            updatable = false
    )
    private UtmComplianceStandardSection section;

    @Column(name = "control_name",
            length = 50
    )
    private String controlName;

    @Column(name = "control_solution")
    private String controlSolution;

    @Column(name = "control_remediation")
    private String controlRemediation;

    @Enumerated(EnumType.STRING)
    @Column(name = "control_strategy")
    private ComplianceStrategy controlStrategy;

    @OneToMany(
            mappedBy = "controlConfig",
            cascade = CascadeType.ALL,
            orphanRemoval = true
    )
    private List<UtmComplianceQueryConfig> queriesConfigs = new ArrayList<>();
}
