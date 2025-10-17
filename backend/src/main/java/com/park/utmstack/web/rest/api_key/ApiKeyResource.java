package com.park.utmstack.web.rest.api_key;

import com.insecureweb.api.domain.User;
import com.insecureweb.api.domain.apikey.index.ApiKeyUsageLogIndexDocument;
import com.insecureweb.api.security.AuthoritiesConstants;
import com.insecureweb.api.security.SecurityUtils;
import com.insecureweb.api.service.ApiKeyService;
import com.insecureweb.api.service.criteria.apikey.ApiKeyUsageLogCriteria;
import com.insecureweb.api.service.dto.SearchHitsResponseDTO;
import com.insecureweb.api.service.dto.apikey.ApiKeyResponseDTO;
import com.insecureweb.api.service.dto.apikey.ApiKeyUpsertDTO;
import com.insecureweb.api.service.exceptions.*;
import com.insecureweb.api.service.user.UserService;
import com.insecureweb.api.util.ResponseUtil;
import com.insecureweb.api.web.rest.restutil.ResponseSearchHitsUtil;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.chart_builder.types.query.FilterType;
import com.park.utmstack.domain.chart_builder.types.query.OperatorType;
import com.park.utmstack.util.ResponseUtil;
import com.park.utmstack.util.UtilPagination;
import io.swagger.v3.oas.annotations.Hidden;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.headers.Header;
import io.swagger.v3.oas.annotations.media.Content;
import io.swagger.v3.oas.annotations.media.Schema;
import io.swagger.v3.oas.annotations.responses.ApiResponse;
import io.swagger.v3.oas.annotations.responses.ApiResponses;
import lombok.AllArgsConstructor;
import org.opensearch.client.opensearch.core.SearchResponse;
import org.opensearch.client.opensearch.core.search.Hit;
import org.opensearch.client.opensearch.core.search.HitsMetadata;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springdoc.core.annotations.ParameterObject;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.access.prepost.PreAuthorize;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.servlet.support.ServletUriComponentsBuilder;
import tech.jhipster.web.util.PaginationUtil;

import java.util.*;

@RestController
@RequestMapping("/api/api-keys")
@PreAuthorize("hasAuthority(\"" + AuthoritiesConstants.USER + "\")")
@AllArgsConstructor
@Hidden
public class ApiKeyResource {

    private static final String CLASSNAME = "ApiKeyResource";
    private final Logger log = LoggerFactory.getLogger(ApiKeyResource.class);

    private final ApiKeyService apiKeyService;
    private final UserService userService;

    private UUID getCurrentAccountId() throws CurrentUserLoginNotFoundException {
        User user = userService.getUserWithAuthoritiesByLogin(SecurityUtils.currentUserLogin());
        return UUID.fromString(user.getAccountId());
    }

    @Operation(summary = "Create API key",
        description = "Creates a new API key record using the provided settings. The plain text key is not generated at creation.")
    @ApiResponses({
        @ApiResponse(responseCode = "201", description = "API key created successfully",
            content = @Content(schema = @Schema(implementation = ApiKeyResponseDTO.class))),
        @ApiResponse(responseCode = "409", description = "API key already exists", content = @Content),
        @ApiResponse(responseCode = "500", description = "Internal server error",
            content = @Content, headers = {
            @Header(name = "X-App-Error", description = "Technical error details")
        })
    })
    @PostMapping
    public ResponseEntity<ApiKeyResponseDTO> createApiKey(@RequestBody ApiKeyUpsertDTO dto) {
        final String ctx = CLASSNAME + ".createApiKey";
        try {
            UUID accountId = getCurrentAccountId();
            ApiKeyResponseDTO responseDTO = apiKeyService.createApiKey(accountId, dto);
            return ResponseEntity.status(HttpStatus.CREATED).body(responseDTO);
        } catch (ApiKeyExistException e) {
            return ResponseUtil.buildConflictResponse(e.getMessage());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg, e);
            return ResponseUtil.buildInternalServerErrorResponse(msg);
        }
    }

    @Operation(summary = "Generate a new API key",
        description = "Generates (or renews) a new random API key for the specified API key record. The plain text key is returned only once.")
    @ApiResponses({
        @ApiResponse(responseCode = "200", description = "API key generated successfully",
            content = @Content(schema = @Schema(type = "string"))),
        @ApiResponse(responseCode = "404", description = "API key not found", content = @Content),
        @ApiResponse(responseCode = "500", description = "Internal server error",
            content = @Content, headers = {
            @Header(name = "X-App-Error", description = "Technical error details")
        })
    })
    @PostMapping("/{id}/generate")
    public ResponseEntity<String> generateApiKey(@PathVariable("id") UUID apiKeyId) {
        final String ctx = CLASSNAME + ".generateApiKey";
        try {
            UUID accountId = getCurrentAccountId();
            String plainKey = apiKeyService.generateApiKey(accountId, apiKeyId);
            return ResponseEntity.ok(plainKey);
        } catch (ApiKeyNotFoundException e) {
            return ResponseUtil.buildNotFoundResponse(e.getMessage());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg, e);
            return ResponseUtil.buildInternalServerErrorResponse(msg);
        }
    }

    @Operation(summary = "Retrieve API key",
        description = "Retrieves the API key details for the specified API key record.")
    @ApiResponses({
        @ApiResponse(responseCode = "200", description = "API key retrieved successfully",
            content = @Content(schema = @Schema(implementation = ApiKeyResponseDTO.class))),
        @ApiResponse(responseCode = "404", description = "API key not found", content = @Content),
        @ApiResponse(responseCode = "500", description = "Internal server error",
            content = @Content, headers = {
            @Header(name = "X-App-Error", description = "Technical error details")
        })
    })
    @GetMapping("/{id}")
    public ResponseEntity<ApiKeyResponseDTO> getApiKey(@PathVariable("id") UUID apiKeyId) {
        final String ctx = CLASSNAME + ".getApiKey";
        try {
            UUID accountId = getCurrentAccountId();
            ApiKeyResponseDTO responseDTO = apiKeyService.getApiKey(accountId, apiKeyId);
            return ResponseEntity.ok(responseDTO);
        } catch (ApiKeyNotFoundException e) {
            return ResponseUtil.buildNotFoundResponse(e.getMessage());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg, e);
            return ResponseUtil.buildInternalServerErrorResponse(msg);
        }
    }

    @Operation(summary = "List API keys",
        description = "Retrieves the API key list.")
    @ApiResponses({
        @ApiResponse(responseCode = "200", description = "API key retrieved successfully",
            content = @Content(schema = @Schema(implementation = ApiKeyResponseDTO.class))),
        @ApiResponse(responseCode = "404", description = "API key not found", content = @Content),
        @ApiResponse(responseCode = "500", description = "Internal server error",
            content = @Content, headers = {
            @Header(name = "X-App-Error", description = "Technical error details")
        })
    })
    @GetMapping("")
    public ResponseEntity<List<ApiKeyResponseDTO>> listApiKeys(@ParameterObject Pageable pageable) {
        final String ctx = CLASSNAME + ".listApiKeys";
        try {
            UUID accountId = getCurrentAccountId();
            Page<ApiKeyResponseDTO> page = apiKeyService.listApiKeys(accountId, pageable);
            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(ServletUriComponentsBuilder.fromCurrentRequest(), page);
            return ResponseEntity.ok().headers(headers).body(page.getContent());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg, e);
            return ResponseUtil.buildInternalServerErrorResponse(msg);
        }
    }

    @Operation(summary = "Update API key",
        description = "Updates mutable fields (name, allowed IPs, expiration) for the specified API key record.")
    @ApiResponses({
        @ApiResponse(responseCode = "200", description = "API key updated successfully",
            content = @Content(schema = @Schema(implementation = ApiKeyResponseDTO.class))),
        @ApiResponse(responseCode = "404", description = "API key not found", content = @Content),
        @ApiResponse(responseCode = "500", description = "Internal server error",
            content = @Content, headers = {
            @Header(name = "X-App-Error", description = "Technical error details")
        })
    })
    @PutMapping("/{id}")
    public ResponseEntity<ApiKeyResponseDTO> updateApiKey(@PathVariable("id") UUID apiKeyId,
                                                          @RequestBody ApiKeyUpsertDTO dto) {
        final String ctx = CLASSNAME + ".updateApiKey";
        try {
            UUID accountId = getCurrentAccountId();
            ApiKeyResponseDTO responseDTO = apiKeyService.updateApiKey(accountId, apiKeyId, dto);
            return ResponseEntity.ok(responseDTO);
        } catch (ApiKeyNotFoundException e) {
            return ResponseUtil.buildNotFoundResponse(e.getMessage());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg, e);
            return ResponseUtil.buildInternalServerErrorResponse(msg);
        }
    }

    @Operation(summary = "Delete API key",
        description = "Deletes the specified API key record for the authenticated user.")
    @ApiResponses({
        @ApiResponse(responseCode = "204", description = "API key deleted successfully", content = @Content),
        @ApiResponse(responseCode = "404", description = "API key not found", content = @Content),
        @ApiResponse(responseCode = "500", description = "Internal server error",
            content = @Content, headers = {
            @Header(name = "X-App-Error", description = "Technical error details")
        })
    })
    @DeleteMapping("/{id}")
    public ResponseEntity<Void> deleteApiKey(@PathVariable("id") UUID apiKeyId) {
        final String ctx = CLASSNAME + ".deleteApiKey";
        try {
            UUID accountId = getCurrentAccountId();
            apiKeyService.deleteApiKey(accountId, apiKeyId);
            return ResponseEntity.noContent().build();
        } catch (ApiKeyNotFoundException e) {
            return ResponseUtil.buildNotFoundResponse(e.getMessage());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg, e);
            return ResponseUtil.buildInternalServerErrorResponse(msg);
        }
    }

    @GetMapping("/usage")
    public ResponseEntity<List<Map>> search(@RequestBody(required = false) List<FilterType> filters,
                                            @RequestParam Integer top, @RequestParam String indexPattern,
                                            @RequestParam(required = false, defaultValue = "false") boolean includeChildren,
                                            Pageable pageable) {
        final String ctx = CLASSNAME + ".search";
        try {
            SearchResponse<Map> searchResponse = elasticsearchService.search(filters, top, indexPattern,
                    pageable, Map.class);

            if (Objects.isNull(searchResponse) || Objects.isNull(searchResponse.hits()) || searchResponse.hits().total().value() == 0)
                return ResponseEntity.ok(Collections.emptyList());

            HitsMetadata<Map> hits = searchResponse.hits();
            HttpHeaders headers = UtilPagination.generatePaginationHttpHeaders(Math.min(hits.total().value(), top),
                    pageable.getPageNumber(), pageable.getPageSize(), "/api/elasticsearch/search");

            return ResponseEntity.ok().headers(headers).body(results);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return com.park.utmstack.util.ResponseUtil.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }
}
