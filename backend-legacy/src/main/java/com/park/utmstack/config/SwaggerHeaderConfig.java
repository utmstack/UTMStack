package com.park.utmstack.config;

import io.swagger.v3.oas.models.parameters.Parameter;
import org.springdoc.core.customizers.OpenApiCustomiser;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import io.swagger.v3.oas.models.media.StringSchema;

@Configuration
public class SwaggerHeaderConfig {

    @Bean
    public OpenApiCustomiser addBypassTFAHeader() {
        return openApi -> openApi.getPaths().values().forEach(pathItem ->
                pathItem.readOperations().forEach(operation -> {
                    Parameter header = new Parameter()
                            .in("header")
                            .name(Constants.TFA_EXEMPTION_HEADER)
                            .required(false)
                            .schema(new StringSchema().example("true"));
                    operation.addParametersItem(header);
                })
        );
    }
}

