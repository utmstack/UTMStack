package com.park.utmstack.domain.application_modules;

import jakarta.persistence.CascadeType;
import jakarta.persistence.OneToMany;
import org.junit.jupiter.api.Test;

import java.lang.reflect.Field;
import java.util.Arrays;
import java.util.List;

import static org.junit.jupiter.api.Assertions.*;

class UtmModuleGroupCascadeTest {

    @Test
    void moduleGroupConfigurationsCascadesRemoveWithoutOrphanRemoval() throws NoSuchFieldException {
        Field field = UtmModuleGroup.class.getDeclaredField("moduleGroupConfigurations");
        OneToMany ann = field.getAnnotation(OneToMany.class);
        assertNotNull(ann, "Expected @OneToMany on UtmModuleGroup.moduleGroupConfigurations");

        List<CascadeType> cascades = Arrays.asList(ann.cascade());
        assertTrue(
            cascades.contains(CascadeType.REMOVE) || cascades.contains(CascadeType.ALL),
            "Must cascade REMOVE so Hibernate 7 deletes eagerly loaded UtmModuleGroupConfiguration children without TransientPropertyValueException"
        );
        assertFalse(
            ann.orphanRemoval(),
            "Must not enable orphanRemoval because PUT /api/utm-configuration-groups can merge groups with an empty configuration set"
        );
        assertFalse(
            cascades.contains(CascadeType.MERGE),
            "Must not cascade MERGE so group metadata updates do not overwrite child configurations"
        );
    }
}
