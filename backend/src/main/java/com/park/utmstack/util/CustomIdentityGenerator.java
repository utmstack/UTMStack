package com.park.utmstack.util;

import org.hibernate.engine.spi.SharedSessionContractImplementor;
import org.hibernate.id.IdentifierGenerator;
import org.hibernate.id.IdentifierGeneratorHelper;

import java.io.Serializable;

/**
 * Identifier generator used by the {@code @GenericGenerator(strategy=...)} on the
 * dashboard / menu / report / correlation / alert-last entities.
 *
 * <p>Behavior (preserved from the Hibernate 5 original, which overrode the
 * before-execution {@code IdentifierGenerator.generate(session, entity)} to return the
 * entity's pre-assigned id when present): the app assigns ids itself before persist —
 * e.g. {@code entity.setId(sequenceNextValue)} (max+1) for dashboards/menus/reports, and
 * {@code UtmAlertLast}'s constructor hard-sets {@code id = 1}. This generator must HONOR
 * that pre-set id; it is a before-execution generator (not a post-insert identity one),
 * so Hibernate keeps the id column in the INSERT and writes the assigned value.
 *
 * <p>Hibernate 6 port: H5's {@code IdentifierGenerator} (before-execution) was split in
 * 6.x into {@link org.hibernate.generator.BeforeExecutionGenerator}. {@link IdentifierGenerator}
 * still implements it, with {@code getEventTypes()} defaulting to {@code INSERT_ONLY}. When
 * no id is pre-set (defensive only — every call site assigns one), it returns
 * {@link IdentifierGeneratorHelper#SHORT_CIRCUIT_INDICATOR}, which defers to the entity's
 * current id / database column default (the sequence-backed columns fill from
 * {@code nextval}, matching the old {@code POST_INSERT_INDICATOR} fallback).
 */
public class CustomIdentityGenerator implements IdentifierGenerator {

    @Override
    public Serializable generate(SharedSessionContractImplementor session, Object entity) {
        // Honor a pre-assigned id (the common path); the persister writes it into the INSERT.
        final Object assigned = session.getEntityPersister(null, entity).getIdentifier(entity, session);
        if (assigned != null) {
            return (Serializable) assigned;
        }
        // No pre-set id: fall through to the database column default (sequence), preserving
        // the original super.generate() -> POST_INSERT behavior for sequence-backed columns.
        return IdentifierGeneratorHelper.SHORT_CIRCUIT_INDICATOR;
    }
}
