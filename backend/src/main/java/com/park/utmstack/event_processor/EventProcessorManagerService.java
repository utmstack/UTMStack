package com.park.utmstack.event_processor;

import com.park.utmstack.config.Constants;
import com.park.utmstack.service.dto.application_modules.ModuleDTO;
import com.park.utmstack.service.web_clients.rest_template.RestTemplateService;
import lombok.RequiredArgsConstructor;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpHeaders;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Service;
import org.springframework.web.util.UriComponentsBuilder;

import java.util.List;

@Service
@RequiredArgsConstructor
public class EventProcessorManagerService {

    private static final String CLASSNAME = "UtmStackConnectionService";
    private final Logger log = LoggerFactory.getLogger(EventProcessorManagerService.class);

    private final RestTemplateService restTemplateService;

    public static final String EVENT_PROCESSOR_BASE_URL = "http://" +
            System.getenv(Constants.ENV_EVENT_PROCESSOR_HOST) + ":" +
            System.getenv(Constants.ENV_EVENT_PROCESSOR_PORT);

    public void updateModule(ModuleDTO module) {
        final String ctx = CLASSNAME + ".updateModule";

        String url = UriComponentsBuilder
                .fromHttpUrl(EVENT_PROCESSOR_BASE_URL + "/api/v1/modules-config")
                .queryParam("nameShort", module.getModuleName())
                .toUriString();

        try{
            ResponseEntity<String> response = restTemplateService.post(
                    url,
                    List.of(module),
                    String.class,
                    buildEventProcessorHeaders()
            );
            response.getStatusCode();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    private HttpHeaders buildEventProcessorHeaders() {
        HttpHeaders headers = new HttpHeaders();
        headers.setContentType(MediaType.APPLICATION_JSON);
        headers.setAccept(List.of(MediaType.ALL));
        headers.set(
                Constants.EVENT_PROCESSOR_INTERNAL_KEY_HEADER,
                System.getenv(Constants.ENV_INTERNAL_KEY)
        );
        return headers;
    }
}
