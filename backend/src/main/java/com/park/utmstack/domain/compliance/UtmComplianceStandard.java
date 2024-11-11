package com.park.utmstack.domain.compliance;

import lombok.Getter;
import lombok.Setter;
import org.hibernate.annotations.GenericGenerator;

import javax.persistence.*;
import java.io.Serializable;

@Setter
@Getter
@Entity
@Table(name = "utm_compliance_standard")
public class UtmComplianceStandard implements Serializable {
    private static final long serialVersionUID = 1L;

    @Id
    @GenericGenerator(name = "CustomIdentityGenerator", strategy = "com.park.utmstack.util.CustomIdentityGenerator")
    @GeneratedValue(generator = "CustomIdentityGenerator")
    private Long id;

    @Column(name = "standard_name")
    private String standardName;

    @Column(name = "standard_description")
    private String standardDescription;

    @Column(name = "system_owner")
    private Boolean systemOwner;

}
