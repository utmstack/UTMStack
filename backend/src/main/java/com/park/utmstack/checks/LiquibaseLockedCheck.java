package com.park.utmstack.checks;

import java.sql.Connection;
import java.sql.ResultSet;
import java.sql.SQLException;
import java.sql.Statement;
import java.util.Objects;

public class LiquibaseLockedCheck {
    private static final String CLASSNAME = "LiquibaseUnlock";

    private final Connection con;

    private LiquibaseLockedCheck(Connection con) {
        if (Objects.isNull(con))
            throw new RuntimeException(CLASSNAME + ": Database connection is null");
        this.con = con;
    }

    public void forceUnlockLiquibase() {
        final String ctx = CLASSNAME + ".forceUnlockLiquibase";
        try {
            ConsoleColors.cyanBold();
            System.out.println(">> Checking liquibase status:");
            if (!existChangeLogLockTable())
                return;
            if (!liquibaseLocked())
                return;
            unlock();
        } catch (Exception e) {
            try {
                con.close();
            } catch (SQLException ex) {
                throw new RuntimeException(ctx + ": " + ex.getLocalizedMessage());
            }
            ConsoleColors.redBold();
            System.out.println(e.getLocalizedMessage());
            ConsoleColors.reset();
            System.exit(1);
        }
    }

    private boolean existChangeLogLockTable() {
        final String ctx = CLASSNAME + ".existChangeLogLockTable";
        ResultSet rs = null;
        try {
            if (!con.isValid(30))
                throw new Exception("Database connection is not valid");
            String query = "SELECT count(*) FROM information_schema.tables WHERE table_name = 'databasechangeloglock' LIMIT 1";
            rs = con.prepareStatement(query).executeQuery();
            boolean exist = rs.next() && rs.getInt(1) == 1;
            if (!exist) {
                ConsoleColors.greenBold();
                System.out.println("\t> First run");
                ConsoleColors.reset();
            }
            rs.close();
            return exist;
        } catch (Exception e) {
            if (!Objects.isNull(rs))
                try {
                    rs.close();
                } catch (SQLException ex) {
                    throw new RuntimeException(ctx + ": " + ex.getLocalizedMessage());
                }
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    private boolean liquibaseLocked() {
        final String ctx = CLASSNAME + ".liquibaseLocked";
        ResultSet rs = null;
        try {
            String query = "select count(d) from databasechangeloglock d where d.\"locked\" is true limit 1";
            rs = con.prepareStatement(query).executeQuery();
            boolean locked = rs.next() && rs.getInt(1) == 1;
            rs.close();
            if (!locked) {
                ConsoleColors.greenBold();
                System.out.println("\t> Success");
                ConsoleColors.reset();
            }
            return locked;
        } catch (Exception e) {
            if (!Objects.isNull(rs))
                try {
                    rs.close();
                } catch (SQLException ex) {
                    throw new RuntimeException(ctx + ": " + ex.getLocalizedMessage());
                }
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    private void unlock() {
        final String ctx = CLASSNAME + ".unlock";
        Statement st = null;
        try {
            ConsoleColors.yellowBold();
            System.out.println("\t> Liquibase is locked, trying to unlock...");
            String query = "UPDATE databasechangeloglock SET \"locked\"=false, lockgranted=null, lockedby=null WHERE id = 1";
            st = con.createStatement();
            st.executeUpdate(query);
            ConsoleColors.greenBold();
            System.out.println("\t> Success");
            ConsoleColors.reset();
            st.close();
        } catch (Exception e) {
            if (!Objects.isNull(st))
                try {
                    st.close();
                } catch (SQLException ex) {
                    throw new RuntimeException(ctx + ": " + ex.getLocalizedMessage());
                }
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    public static LiquibaseLockedCheck getInstance(Connection con) {
        return new LiquibaseLockedCheck(con);
    }
}
