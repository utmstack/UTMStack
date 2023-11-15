package com.utmstack.userauditor;

import com.utmstack.userauditor.checks.DatabaseConnectionCheck;
import com.utmstack.userauditor.checks.ElasticsearchConnectionCheck;
import com.utmstack.userauditor.checks.LiquibaseLockedCheck;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.env.EnvironmentPostProcessor;
import org.springframework.core.env.ConfigurableEnvironment;

import java.sql.Connection;
import java.sql.SQLException;
import java.util.Objects;

public class ApplicationStartProcessor implements EnvironmentPostProcessor {
    private static final String CLASSNAME = "ApplicationStartProcessor";
    private Connection con;

    @Override
    public void postProcessEnvironment(ConfigurableEnvironment environment, SpringApplication application) {
        final String ctx = CLASSNAME + ".postProcessEnvironment";
        try {
            System.out.println("------------------------------------------------");
            System.out.println("- Checking application start requirements");
            System.out.println("------------------------------------------------");
            // Checking database connection before application start
            databaseConnectionCheck();
            System.out.println("\t> Success");

            // Checking if liquibase is locked
            liquibaseLockedCheck(con);

            // Checking elasticsearch connection
            elasticsearchConnectionCheck();
        } catch (Exception e) {
            if (!Objects.isNull(con)) {
                try {
                    con.close();
                } catch (SQLException ex) {
                    throw new RuntimeException(ctx + ": " + ex.getLocalizedMessage());
                }
            }
            System.out.println("\t> " + e.getLocalizedMessage());
            System.exit(1);
        }
    }

    private void databaseConnectionCheck() {
        con = DatabaseConnectionCheck.getInstance().databaseConnectionCheck();
    }

    private void liquibaseLockedCheck(Connection con) {
        LiquibaseLockedCheck.getInstance(con).forceUnlockLiquibase();
    }

    private void elasticsearchConnectionCheck() {
        ElasticsearchConnectionCheck.getInstance().connectionCheck(5);
    }
}
