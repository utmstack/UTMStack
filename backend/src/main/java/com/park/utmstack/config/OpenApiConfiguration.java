package com.park.utmstack.config;

import com.park.utmstack.util.MapUtil;
import io.swagger.v3.oas.models.Components;
import io.swagger.v3.oas.models.OpenAPI;
import io.swagger.v3.oas.models.info.Info;
import io.swagger.v3.oas.models.security.SecurityRequirement;
import io.swagger.v3.oas.models.security.SecurityScheme;
import io.swagger.v3.oas.models.servers.Server;
import org.springframework.boot.actuate.info.InfoEndpoint;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class OpenApiConfiguration {

    private final InfoEndpoint infoEndpoint;

    public OpenApiConfiguration(InfoEndpoint infoEndpoint) {
        this.infoEndpoint = infoEndpoint;
    }

    @Bean
    public OpenAPI customOpenAPI() {
        final String securitySchemeBearer = "bearerAuth";
        final String securitySchemeApiKey = "ApiKeyAuth";
        final String apiTitle = "UTMStack API";
        String version = MapUtil.flattenToStringMap(infoEndpoint.info(), true).get("build.version");
        return new OpenAPI()
            .addSecurityItem(new SecurityRequirement().addList(securitySchemeBearer).addList(securitySchemeApiKey))
            .components(new Components()
                .addSecuritySchemes(securitySchemeBearer,
                    new SecurityScheme()
                        .name(securitySchemeBearer)
                        .type(SecurityScheme.Type.HTTP)
                        .scheme("bearer")
                        .bearerFormat("JWT"))
                .addSecuritySchemes(securitySchemeApiKey, new SecurityScheme()
                    .name("Utm-Internal-Key")
                    .type(SecurityScheme.Type.APIKEY)
                    .in(SecurityScheme.In.HEADER)))
            .info(new Info().title(apiTitle).version(version))
            .addServersItem(new Server().url("/").description("Default Server URL"));
    }
}
