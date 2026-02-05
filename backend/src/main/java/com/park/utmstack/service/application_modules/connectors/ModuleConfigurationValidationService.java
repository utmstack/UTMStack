package com.park.utmstack.service.application_modules.connectors;

import com.fasterxml.jackson.databind.JsonNode;
import com.park.utmstack.config.Constants;
import com.park.utmstack.service.dto.application_modules.ModuleConfigValidationErrorMapper;
import com.park.utmstack.service.dto.application_modules.ModuleConfigValidationErrorResponse;
import com.park.utmstack.service.dto.application_modules.UtmModuleGroupConfWrapperDTO;
import com.park.utmstack.service.web_clients.rest_template.RestTemplateService;
import com.park.utmstack.util.exceptions.ApiException;
import lombok.RequiredArgsConstructor;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.*;
import org.springframework.stereotype.Service;

import java.util.List;


@Service
@RequiredArgsConstructor
public class ModuleConfigurationValidationService {

    private final Logger log = LoggerFactory.getLogger(ModuleConfigurationValidationService.class);
    private final RestTemplateService restTemplateService;
    private static final String CLASSNAME = "UtmStackConnectionService";


    public boolean validateModuleConfiguration(String module, UtmModuleGroupConfWrapperDTO configurations) {
        final String ctx = CLASSNAME + ".ModuleConfigurationValidationService";

        HttpHeaders headers = new HttpHeaders();
        headers.add("Content-Type", "application/json");
        headers.add("Accept", "*/*");
        headers.set(Constants.EVENT_PROCESSOR_INTERNAL_KEY_HEADER, System.getenv(Constants.ENV_INTERNAL_KEY));

        String baseUrl = "http://" + System.getenv(Constants.ENV_EVENT_PROCESSOR_HOST) + ":" + System.getenv(Constants.ENV_EVENT_PROCESSOR_PORT);
        String endPoint = baseUrl + "/api/v1/modules-config/validate?nameShort=" + module;

        ResponseEntity<JsonNode> response = restTemplateService.postRaw(
                endPoint,
                configurations,
                JsonNode.class,
                headers
        );

        JsonNode body = response.getBody();

        if (response.getStatusCode().is2xxSuccessful() && body != null && body.has("status")) {
            return true;
        }

        if (body != null && body.has("error")) {
            String errorText = body.get("error").asText();

            if (errorText.contains("{\"meta\"")) {
                ModuleConfigValidationErrorResponse structured = ModuleConfigValidationErrorMapper.parse(errorText);

                if (structured != null) {
                    String traceId = structured.getMeta().getTraceId();
                    String message = structured.getErrors().get(0).getMessage();

                    log.error("{}: External provider validation failed for module {}. TraceId: {}. Message: {}",
                            ctx, module, traceId, message);

                    throw new ApiException(
                            "External provider validation failed: " + message + " (traceId=" + traceId + ")",
                            HttpStatus.UNAUTHORIZED
                    );
                }
            }

            log.error("{}: Module configuration validation failed for module {}. Cause: {}",
                    ctx, module, errorText);

            throw new ApiException(errorText, HttpStatus.BAD_REQUEST);
        }

        log.error("{}: Unexpected response validating module {}.", ctx, module);
        throw new ApiException(
                String.format("%s: Unexpected response validating module %s.", ctx, module),
                HttpStatus.INTERNAL_SERVER_ERROR
        );
    }


}

