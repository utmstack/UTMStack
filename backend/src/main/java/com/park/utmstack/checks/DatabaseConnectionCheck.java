package com.park.utmstack.checks;

import com.park.utmstack.config.Constants;
import org.springframework.boot.jdbc.DataSourceBuilder;

import java.sql.Connection;

public class DatabaseConnectionCheck {
    private static final String CLASSNAME = "DatabaseConnectionCheck";

    /**
     * Build a datasource and open a connection
     *
     * @return A database connection
     */
    public Connection databaseConnectionCheck() {
        final String ctx = CLASSNAME + ".databaseConnectionCheck";
        short retry = 5;
        ConsoleColors.cyanBold();
        System.out.println(">> Checking database connection:");
        while (retry > 0) {
            try {
                String dbUsername = System.getenv(Constants.ENV_DB_USER);
                String dbPassword = System.getenv(Constants.ENV_DB_PASS);
                String dbHost = System.getenv(Constants.ENV_DB_HOST);
                String dbPort = System.getenv(Constants.ENV_DB_PORT);
                String dbName = System.getenv(Constants.ENV_DB_NAME);
                String driver = "org.postgresql.Driver";

                String connectionUrl = String.format("jdbc:postgresql://%1$s:%2$s/%3$s", dbHost, dbPort, dbName);

                Connection connection = DataSourceBuilder.create().username(dbUsername).password(dbPassword)
                    .url(connectionUrl).driverClassName(driver).build().getConnection();

                ConsoleColors.greenBold();
                System.out.println("\t> Success");
                ConsoleColors.reset();

                return connection;
            } catch (Exception e) {
                ConsoleColors.redBold();
                System.out.println("\t> " + e.getLocalizedMessage());
                ConsoleColors.reset();
                retry--;

                ConsoleColors.yellowBold();
                for (int i = 10; i > 0; i--) {
                    System.out.printf("\t> Retrying in: %1$s\r", i);
                    try {
                        Thread.sleep(1000L);
                    } catch (Exception ex) {
                        throw new RuntimeException(ctx + ": " + ex.getLocalizedMessage());
                    }
                }
            }
        }
        ConsoleColors.reset();
        throw new RuntimeException("Fail to establish connection with database");
    }

    public static DatabaseConnectionCheck getInstance() {
        return new DatabaseConnectionCheck();
    }
}
