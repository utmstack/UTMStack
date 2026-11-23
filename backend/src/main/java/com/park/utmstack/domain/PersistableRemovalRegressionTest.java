package com.park.utmstack.domain;

import com.park.utmstack.domain.alert_response_rule.UtmAlertResponseRule;
import com.park.utmstack.domain.chart_builder.UtmDashboard;
import com.park.utmstack.domain.chart_builder.UtmVisualization;
import com.park.utmstack.domain.collector.UtmCollector;
import com.park.utmstack.domain.compliance.UtmComplianceReportConfig;
import com.park.utmstack.domain.compliance.UtmComplianceStandardSection;
import com.park.utmstack.domain.correlation.config.UtmDataTypes;
import com.park.utmstack.domain.correlation.config.UtmRegexPattern;
import com.park.utmstack.domain.getting_started.UtmGettingStarted;
import com.park.utmstack.domain.index_pattern.UtmIndexPattern;
import com.park.utmstack.domain.logstash_filter.UtmLogstashFilter;
import com.park.utmstack.domain.logstash_filter.UtmLogstashFilterGroup;
import com.park.utmstack.domain.reports.UtmReport;
import com.park.utmstack.domain.reports.UtmReportSection;
import jakarta.persistence.GeneratedValue;
import org.junit.jupiter.api.Test;
import org.springframework.data.domain.Persistable;

import java.lang.reflect.Field;
import java.util.List;

import static org.junit.jupiter.api.Assertions.*;

class PersistableRemovalRegressionTest {

    @Test
    void entitiesDoNotImplementPersistableSoDtoUpdatesMergeCleanly() {
        List<Class<?>> entities = List.of(
            UtmMenu.class,
            UtmSpaceNotificationControl.class,
            UtmAlertResponseRule.class,
            UtmDashboard.class,
            UtmVisualization.class,
            UtmCollector.class,
            UtmComplianceReportConfig.class,
            UtmComplianceStandardSection.class,
            UtmDataTypes.class,
            UtmRegexPattern.class,
            UtmGettingStarted.class,
            UtmIndexPattern.class,
            UtmLogstashFilter.class,
            UtmLogstashFilterGroup.class,
            UtmReport.class,
            UtmReportSection.class
        );
        for (Class<?> cls : entities) {
            assertFalse(Persistable.class.isAssignableFrom(cls),
                cls.getSimpleName() + " must not implement Persistable (breaks Hibernate 7 PUT merge on detached DTOs)");
        }
    }

    @Test
    void assignedIdEntitiesHaveNoGeneratedValueAnnotation() throws NoSuchFieldException {
        for (Class<?> cls : List.of(UtmAlertLast.class, UtmSpaceNotificationControl.class, UtmCollector.class)) {
            Field id = cls.getDeclaredField("id");
            assertNull(id.getAnnotation(GeneratedValue.class),
                cls.getSimpleName() + ".id must be a plain assigned @Id so merge() upserts cleanly under Hibernate 7");
        }
    }
}
