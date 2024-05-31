package com.park.utmstack.web.rest.user_auditor;

import com.park.utmstack.config.Constants;
import com.park.utmstack.service.dto.user_auditor.UtmAuditorUsersDTO;
import com.park.utmstack.util.UtilResponse;
import com.park.utmstack.web.rest.user_auditor.dto.MicroserviceRequest;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.core.ParameterizedTypeReference;
import org.springframework.data.domain.Pageable;
import org.springframework.http.*;
import org.springframework.util.StringUtils;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.client.RestTemplate;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * REST controller for managing {@link UtmAuditorUsersDTO}.
 */
@RestController
@RequestMapping("/api")
public class UtmAuditorUsersResource {

    private final Logger log = LoggerFactory.getLogger(UtmAuditorUsersResource.class);

    private String microServiceUrl;

    private static final String CLASSNAME = "UserResource";

    public UtmAuditorUsersResource() {

        if (!StringUtils.hasText(System.getenv(Constants.ENV_AD_AUDIT_SERVICE))) {
            this.microServiceUrl = "http://user-auditor:8080/api";
        } else{
            this.microServiceUrl = System.getenv(Constants.ENV_AD_AUDIT_SERVICE);
        }
    }


    @GetMapping("/winlogbeat-info-by-filter")
    public ResponseEntity<List<Map>> getUtmAuditorUsers(@RequestParam String sid,
                                                        @RequestParam String from,
                                                        @RequestParam String to,
                                                        @RequestParam String indexPattern,
                                                        @RequestParam String sort,
                                                        @RequestParam String page,
                                                        @RequestParam String size,
                                                        Pageable pageable) {


        final String ctx = CLASSNAME + ".search";
        try {

            Map<String, String> urlParams = new HashMap<>();
            urlParams.put("sid", sid);
            urlParams.put("from", from);
            urlParams.put("to", to);
            urlParams.put("indexPattern", indexPattern);
            urlParams.put("sort", sort);
            urlParams.put("page", page);
            urlParams.put("size", size);


            RestTemplate restTemplate = new RestTemplate();

            HttpHeaders headers = new HttpHeaders();
            headers.setContentType(MediaType.APPLICATION_JSON);

            MicroserviceRequest request = new MicroserviceRequest();

            HttpEntity<MicroserviceRequest> requestEntity = new HttpEntity<>(request, headers);

            String uri = microServiceUrl.concat("/search?page={page}&size={size}&sid={sid}&from={from}&to={to}&indexPattern={indexPattern}&sort={sort}");

            return restTemplate.exchange(
                uri,
                HttpMethod.GET,
                requestEntity,
                new ParameterizedTypeReference<List<Map>>() {
                },
                urlParams
            );
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }

    }

    /**
     * {@code GET  /utm-auditor-users} : get all the utmAuditorUsers.
     *
     * @param pageable the pagination information.
     * @return the {@link ResponseEntity} with status {@code 200 (OK)} and the list of utmAuditorUsers in body.
     */
    @GetMapping("/utm-auditor-users-by-src")
    public ResponseEntity<List<UtmAuditorUsersDTO>> getAllUtmAuditorUsers(Pageable pageable, Long sourceId) {

        final String ctx = CLASSNAME + ".search";
        try {
            Map<String, String> urlParams = new HashMap<>();
            urlParams.put("page", String.valueOf(pageable.getPageNumber()));
            urlParams.put("size", String.valueOf(pageable.getPageSize()));
            urlParams.put("id", String.valueOf(sourceId));
            urlParams.put("sort", "name,asc");

            RestTemplate restTemplate = new RestTemplate();

            String uri = microServiceUrl.concat("/utm-auditor-users-by-src?page={page}&size={size}&sort={sort}&id={id}");

            HttpHeaders headers = new HttpHeaders();
            headers.setContentType(MediaType.APPLICATION_JSON);

            MicroserviceRequest request = new MicroserviceRequest();
            request.setId(sourceId);
            request.setPageable(pageable);

            HttpEntity<MicroserviceRequest> requestEntity = new HttpEntity<>(request, headers);

            return restTemplate.exchange(
                uri,
                HttpMethod.GET,
                requestEntity,
                new ParameterizedTypeReference<List<UtmAuditorUsersDTO>>() {
                },
                urlParams
            );
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }
}
