package com.park.utmstack.domain.compliance;

import com.park.utmstack.domain.compliance.enums.EvaluationRule;
import com.park.utmstack.domain.index_pattern.UtmIndexPattern;
import lombok.Getter;
import lombok.Setter;
import org.hibernate.annotations.GenericGenerator;

import javax.persistence.*;
import javax.validation.constraints.Size;
import java.io.Serializable;

@Getter
@Setter
@Entity
@Table(name = "utm_compliance_query_config")
public class UtmComplianceQueryConfig implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GenericGenerator(
            name = "CustomIdentityGenerator",
            strategy = "com.park.utmstack.util.CustomIdentityGenerator"
    )
    @GeneratedValue(generator = "CustomIdentityGenerator")
    private Long id;

    @Column(name = "query_name",
            nullable = false,
            columnDefinition = "TEXT"
    )
    @Size(min = 10, max = 200)
    private String queryName;

    @Column(name = "query_description",
            length = 2000,
            nullable = false,
            columnDefinition = "TEXT"
    )
    private String queryDescription;

    @Column(name = "sql_query",
            length = 2000,
            nullable = false,
            columnDefinition = "TEXT"
    )
    private String sqlQuery;

    @Enumerated(EnumType.STRING)
    @Column(name = "evaluation_rule",
            nullable = false
    )
    private EvaluationRule evaluationRule;

    private Integer ruleValue;

    @Column(name = "index_pattern_id",
            nullable = false
    )
    private Long indexPatternId;

    @ManyToOne
    @JoinColumn(name = "index_pattern_id",
            insertable = false,
            updatable = false
    )
    private UtmIndexPattern indexPattern;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "control_config_id", nullable = false)
    private UtmComplianceControlConfig controlConfig;
}
