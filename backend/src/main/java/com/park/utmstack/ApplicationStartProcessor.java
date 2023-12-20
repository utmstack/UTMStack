package com.park.utmstack;


import com.park.utmstack.checks.*;
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
            ConsoleColors.magentaBold();
            System.out.println("------------------------------------------------");
            System.out.println("- Checking application start requirements");
            System.out.println("------------------------------------------------");
            ConsoleColors.reset();

            // Checking environment variables
            environmentVariablesCheck();

            // Checking database connection before application start
            databaseConnectionCheck();

            // Checking if liquibase is locked
            liquibaseLockedCheck(con);

            // Checking elasticsearch connection
            elasticsearchConnectionCheck();

            ConsoleColors.magentaBold();
            System.out.println("------------------------------------------------");
            ConsoleColors.reset();
        } catch (Exception e) {
            if (!Objects.isNull(con)) {
                try {
                    con.close();
                } catch (SQLException ex) {
                    throw new RuntimeException(ctx + ": " + ex.getLocalizedMessage());
                }
            }
            ConsoleColors.redBold();
            System.out.println("\t> " + e.getLocalizedMessage() + "\n");
            ConsoleColors.reset();
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

    private void environmentVariablesCheck() {
        EnvironmentVariablesCheck.getInstance().checkEnvironmentVariables();
    }
}
