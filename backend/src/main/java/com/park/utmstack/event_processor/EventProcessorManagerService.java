package com.park.utmstack.event_processor;

import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.application_modules.UtmModule;
import com.park.utmstack.domain.application_modules.UtmModuleGroup;
import com.park.utmstack.domain.application_modules.enums.ModuleName;
import com.park.utmstack.service.web_clients.rest_template.RestTemplateService;
import com.park.utmstack.util.CipherUtil;
import lombok.RequiredArgsConstructor;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpHeaders;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Service;
import org.springframework.util.StringUtils;
import org.springframework.web.util.UriComponentsBuilder;

import java.util.List;
import java.util.Set;

@Service
@RequiredArgsConstructor
public class EventProcessorManagerService {

    private static final String CLASSNAME = "UtmStackConnectionService";
    private final Logger log = LoggerFactory.getLogger(EventProcessorManagerService.class);

    private final RestTemplateService restTemplateService;

    private final List<ModuleName> typeFileNeedsDecryptList = List.of(ModuleName.GCP);

    public static final String EVENT_PROCESSOR_BASE_URL = "http://" +
            System.getenv(Constants.ENV_EVENT_PROCESSOR_HOST) + ":" +
            System.getenv(Constants.ENV_EVENT_PROCESSOR_PORT);

    public void updateModule(UtmModule module) {
        final String ctx = CLASSNAME + ".updateModule";

        String url = UriComponentsBuilder
                .fromHttpUrl(EVENT_PROCESSOR_BASE_URL + "/api/v1/modules-config")
                .queryParam("nameShort", module.getModuleName())
                .toUriString();

        try{
            this.decryptModuleConfig (module);
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

    public void decryptModuleConfig (UtmModule module){
        Set<UtmModuleGroup> groups = module.getModuleGroups();
        groups.forEach((gp) -> {
            gp.getModuleGroupConfigurations().forEach((gpc) -> {
                if ((gpc.getConfDataType().equals(Constants.CONF_TYPE_PASSWORD) && StringUtils.hasText(gpc.getConfValue()))
                        || (gpc.getConfDataType().equals(Constants.CONF_TYPE_FILE) && StringUtils.hasText(gpc.getConfValue())) && typeFileNeedsDecryptList.contains(module.getModuleName())) {
                    gpc.setConfValue(CipherUtil.decrypt(gpc.getConfValue(), System.getenv(Constants.ENV_ENCRYPTION_KEY)));
                }
            });
        });
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
