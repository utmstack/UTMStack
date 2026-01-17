package com.park.utmstack.domain.compliance;

import com.park.utmstack.domain.compliance.enums.EvaluationRule;
import com.park.utmstack.domain.index_pattern.UtmIndexPattern;
import lombok.Getter;
import lombok.Setter;
import org.hibernate.annotations.GenericGenerator;

import javax.persistence.*;
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

    @Column(name = "query_description",
            columnDefinition = "TEXT"
    )
    private String queryDescription;

    @Column(name = "sql_query",
            columnDefinition = "TEXT"
    )
    private String sqlQuery;

    @Enumerated(EnumType.STRING)
    @Column(name = "evaluation_rule")
    private EvaluationRule evaluationRule;

    @Column(name = "index_pattern_id")
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
