/*
 * Vendored from tech.jhipster:JHipster 7.3.1 (tech/jhipster/domain/util/FixedPostgreSQL10Dialect.java)
 * and adapted for Hibernate 6.2 (Spring Boot 3.1.5). JHipster 8 removed this class.
 *
 * The Hibernate 5 original overrode remapSqlTypeDescriptor to force Types.BLOB -> bytea.
 * Hibernate 6 removed the org.hibernate.type.descriptor.sql.* API that override targeted, and
 * its PostgreSQL dialects already map binary large objects to bytea, so the Hibernate-6 form is
 * an empty subclass that keeps the stable class name referenced by
 * spring.jpa.properties.hibernate.dialect / database-platform in application-dev.yml and
 * application-prod.yml. Behavior for the app's bytea columns is unchanged.
 */
package tech.jhipster.domain.util;

import org.hibernate.dialect.PostgreSQLDialect;

/**
 * Hibernate 6.6 removed the version-specific {@code PostgreSQL10Dialect} (all
 * PostgreSQL versions now share {@link PostgreSQLDialect}), so this class — kept only
 * so the stable FQN referenced by {@code spring.jpa.properties.hibernate.dialect} in
 * application-dev.yml / application-prod.yml still resolves — extends the single 6.6 class.
 */
public class FixedPostgreSQL10Dialect extends PostgreSQLDialect {
}
