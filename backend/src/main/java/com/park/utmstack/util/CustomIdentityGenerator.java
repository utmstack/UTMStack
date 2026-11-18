package com.park.utmstack.util;

import org.hibernate.engine.spi.SharedSessionContractImplementor;
import org.hibernate.id.IdentifierGenerator;
import org.hibernate.jdbc.ReturningWork;

import jakarta.persistence.Table;
import java.io.Serializable;
import java.lang.reflect.Field;
import java.sql.Connection;
import java.sql.PreparedStatement;
import java.sql.ResultSet;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

/**
 * Identifier generator used by the {@code @GenericGenerator(strategy=...)} on the
 * dashboard / menu / report / correlation / ... entities (22 total).
 *
 * <p>Contract (restored from the Hibernate 5 original, which extended
 * {@code IdentityGenerator} and overrode {@code generate(session, entity)} to return the
 * entity's pre-assigned id when present):
 * <ul>
 *   <li>If the app pre-assigned an id ({@code entity.setId(...)} — the dev/system
 *       "next system value" path, or {@code UtmAlertLast}'s hard-set {@code id=1}),
 *       HONOR it: return that value so Hibernate writes it into the INSERT.</li>
 *   <li>If no id was pre-assigned (the prod user-create path), fetch the next value from
 *       the entity's id sequence and return it, so the INSERT carries a real, unique id.</li>
 * </ul>
 *
 * <p>Hibernate 7 port — why this differs from the migration version: the migration (#3131)
 * rewrote this to return {@code IdentifierGeneratorHelper.SHORT_CIRCUIT_INDICATOR} for the
 * no-pre-set-id case. That is incompatible with Hibernate 7.4: in
 * {@code AbstractSaveEventListener.saveWithGeneratedId}, a before-execution generator's
 * {@code SHORT_CIRCUIT_INDICATOR} return routes Hibernate into
 * {@code session.getIdentifier(entity)}, which throws
 * {@code "Given entity is not associated with the persistence context"} for a freshly
 * {@code persist()}ed (transient) entity — breaking EVERY prod create of these entities.
 * Returning a real, sequence-derived id instead takes Hibernate's
 * {@code persister.setIdentifier(entity, id)} branch (which works for transient entities),
 * which is value-equivalent to letting the id column's {@code DEFAULT nextval(...)} fill it.
 *
 * <p>The pre-set id is read reflectively (the {@code id} field) — no persistence-context or
 * entity-name dependency. The sequence is resolved from the id column's
 * {@code DEFAULT nextval('...')} (cached per entity class), falling back to
 * {@code <table>_id_seq}; it always advances by exactly one, and the companion changeset
 * {@code 20261118001_sync_custom_identity_sequences} positions every sequence at or above
 * its {@code max(id)} (respecting the reserved >=1000000 user ranges) so no value can collide.
 */
public class CustomIdentityGenerator implements IdentifierGenerator {

    private static final Map<Class<?>, Field> ID_FIELDS = new ConcurrentHashMap<>();
    private static final Map<Class<?>, String> SEQ_BY_ENTITY = new ConcurrentHashMap<>();
    private static final Map<String, String> SEQ_BY_TABLE = new ConcurrentHashMap<>();
    private static final Pattern NEXTVAL = Pattern.compile("nextval\\(['\"]?\\s*([A-Za-z0-9_]+)\\s*['\"]?");

    @Override
    public Serializable generate(SharedSessionContractImplementor session, Object entity) {
        final Field idField = findIdField(entity.getClass());
        final Object preset = (idField != null) ? readField(entity, idField) : null;
        if (preset != null) {
            return (Serializable) preset;
        }
        // No pre-set id: fetch the next sequence value and return it as a real id.
        final String seq = sequenceFor(entity.getClass(), session);
        return session.doReturningWork((ReturningWork<Long>) conn -> nextVal(conn, seq));
    }

    private static String sequenceFor(Class<?> entityClass, SharedSessionContractImplementor session) {
        String cached = SEQ_BY_ENTITY.get(entityClass);
        if (cached != null) {
            return cached;
        }
        String table = tableName(entityClass);
        String seq = SEQ_BY_TABLE.get(table);
        if (seq == null) {
            seq = session.doReturningWork((ReturningWork<String>) conn -> {
                try (PreparedStatement ps = conn.prepareStatement(
                        "SELECT column_default FROM information_schema.columns "
                                + "WHERE table_name = ? AND column_name = 'id'")) {
                    ps.setString(1, table);
                    try (ResultSet rs = ps.executeQuery()) {
                        if (rs.next()) {
                            String def = rs.getString(1);
                            if (def != null) {
                                Matcher m = NEXTVAL.matcher(def);
                                if (m.find()) {
                                    return m.group(1);
                                }
                            }
                        }
                    }
                }
                return null;
            });
            if (seq == null) {
                seq = table + "_id_seq";
            }
            SEQ_BY_TABLE.put(table, seq);
        }
        SEQ_BY_ENTITY.put(entityClass, seq);
        return seq;
    }

    private static String tableName(Class<?> entityClass) {
        Table t = entityClass.getAnnotation(Table.class);
        if (t != null && t.name() != null && !t.name().isEmpty()) {
            return t.name();
        }
        // Fallback: entity simple name (matches the convention used by these tables).
        return entityClass.getSimpleName();
    }

    private static Field findIdField(Class<?> entityClass) {
        Field cached = ID_FIELDS.get(entityClass);
        if (cached != null) {
            return cached;
        }
        Class<?> c = entityClass;
        while (c != null && c != Object.class) {
            try {
                Field f = c.getDeclaredField("id");
                f.setAccessible(true);
                ID_FIELDS.put(entityClass, f);
                return f;
            } catch (NoSuchFieldException nsfe) {
                c = c.getSuperclass();
            }
        }
        ID_FIELDS.put(entityClass, null);
        return null;
    }

    private static Object readField(Object entity, Field field) {
        try {
            return field.get(entity);
        } catch (IllegalAccessException iae) {
            throw new IllegalStateException("CustomIdentityGenerator: cannot read id field " + field, iae);
        }
    }

    private static Long nextVal(Connection conn, String seq) throws java.sql.SQLException {
        try (PreparedStatement ps = conn.prepareStatement("SELECT nextval(?)::bigint")) {
            ps.setString(1, seq);
            try (ResultSet rs = ps.executeQuery()) {
                if (rs.next()) {
                    return rs.getLong(1);
                }
            }
        }
        throw new IllegalStateException("CustomIdentityGenerator: nextval('" + seq + "') returned no value");
    }
}
