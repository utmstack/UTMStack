package com.park.utmstack.checks;

import com.park.utmstack.config.Constants;
import org.springframework.util.CollectionUtils;
import org.springframework.util.StringUtils;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;

public class EnvironmentVariablesCheck {

    public void checkEnvironmentVariables() {
        ConsoleColors.cyanBold();
        System.out.println(">> Checking environment variables:");
        Map<String, String> vars = System.getenv();

        List<String> missingVars = new ArrayList<>();

        if (!StringUtils.hasText(vars.get(Constants.ENV_DB_HOST)))
            missingVars.add(Constants.ENV_DB_HOST);
        if (!StringUtils.hasText(vars.get(Constants.ENV_DB_PORT)))
            missingVars.add(Constants.ENV_DB_PORT);
        if (!StringUtils.hasText(vars.get(Constants.ENV_DB_NAME)))
            missingVars.add(Constants.ENV_DB_NAME);
        if (!StringUtils.hasText(vars.get(Constants.ENV_DB_USER)))
            missingVars.add(Constants.ENV_DB_USER);
        if (!StringUtils.hasText(vars.get(Constants.ENV_DB_PASS)))
            missingVars.add(Constants.ENV_DB_PASS);
        if (!StringUtils.hasText(vars.get(Constants.ENV_ELASTICSEARCH_HOST)))
            missingVars.add(Constants.ENV_ELASTICSEARCH_HOST);
        if (!StringUtils.hasText(vars.get(Constants.ENV_ELASTICSEARCH_PORT)))
            missingVars.add(Constants.ENV_ELASTICSEARCH_PORT);
        if (!StringUtils.hasText(vars.get(Constants.ENV_GRPC_AGENT_MANAGER_HOST)))
            missingVars.add(Constants.ENV_GRPC_AGENT_MANAGER_HOST);
        if (!StringUtils.hasText(vars.get(Constants.ENV_GRPC_AGENT_MANAGER_PORT)))
            missingVars.add(Constants.ENV_GRPC_AGENT_MANAGER_PORT);
        if (!StringUtils.hasText(vars.get(Constants.ENV_INTERNAL_KEY)))
            missingVars.add(Constants.ENV_INTERNAL_KEY);
        if (!StringUtils.hasText(vars.get(Constants.ENV_ENCRYPTION_KEY)))
            missingVars.add(Constants.ENV_ENCRYPTION_KEY);
        if (!StringUtils.hasText(vars.get(Constants.ENV_SERVER_NAME)))
            missingVars.add(Constants.ENV_SERVER_NAME);

        if (CollectionUtils.isEmpty(missingVars)) {
            ConsoleColors.greenBold();
            System.out.println("\t> Success");
            ConsoleColors.reset();
            return;
        }
        ConsoleColors.redBold();
        System.out.printf("\t > Missing Variables: [%1$s]\n\n", String.join(", ", missingVars));
        ConsoleColors.reset();
        System.exit(1);
    }

    public static EnvironmentVariablesCheck getInstance() {
        return new EnvironmentVariablesCheck();
    }
}
