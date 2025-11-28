package com.park.utmstack.service.application_modules.connectors;

import com.park.utmstack.config.Constants;
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

        String baseUrl = "http://" + System.getenv(Constants.ENV_EVENT_PROCESSOR_HOST)  + ":" + System.getenv(Constants.ENV_EVENT_PROCESSOR_PORT);
        String endPoint = baseUrl + "/api/v1/modules-config/validate?nameShort=" + module;
        try{
            ResponseEntity<String> response = restTemplateService.post(
                    endPoint,
                    configurations,
                    String.class,
                    headers
            );

            if (!response.getStatusCode().is2xxSuccessful()) {
                List<String> errors = response.getHeaders().get("X-UtmStack-error");
                String errorMessage = (errors != null && !errors.isEmpty())
                        ? String.join(", ", errors)
                        : "Unknown error occurred during module configuration validation.";

                log.error("{}: Module configuration validation failed for module: {} with status: {}. Cause: {}",
                        ctx, module, response.getStatusCode(), errorMessage);
                throw new ApiException(
                        String.format("Module configuration validation failed for module: %s. Cause: %s", module, errorMessage),
                        response.getStatusCode()
                );
            }

            return  true;

        } catch (ApiException e) {
            throw e;
        } catch (Exception e) {
            log.error("{}: An error occurred while validating module configuration for module: {}. Cause: {}",
                    ctx, module, e.getMessage(), e);
            throw new ApiException("An error occurred while validating module configuration", HttpStatus.INTERNAL_SERVER_ERROR);
        }
    }
}

