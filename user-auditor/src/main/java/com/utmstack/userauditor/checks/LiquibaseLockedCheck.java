package com.utmstack.userauditor.checks;

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
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
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
            if (!exist)
                System.out.println("\t> First run");
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
            if (!locked)
                System.out.println("\t> Unlocked");
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
            System.out.println("\t> Locked");
            System.out.println("\t\t> Unlocking...");
            String query = "UPDATE databasechangeloglock SET \"locked\"=false, lockgranted=null, lockedby=null WHERE id = 1";
            st = con.createStatement();
            st.executeUpdate(query);
            System.out.println("\t> Unlocked");
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
