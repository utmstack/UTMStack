package com.park.utmstack.web.rest.api_key;


import com.park.utmstack.domain.chart_builder.types.query.FilterType;
import com.park.utmstack.security.AuthoritiesConstants;
import com.park.utmstack.service.UserService;
import com.park.utmstack.service.api_key.ApiKeyService;
import com.park.utmstack.service.dto.api_key.ApiKeyResponseDTO;
import com.park.utmstack.service.dto.api_key.ApiKeyUpsertDTO;
import com.park.utmstack.service.elasticsearch.ElasticsearchService;
import com.park.utmstack.util.UtilPagination;
import com.park.utmstack.web.rest.elasticsearch.ElasticsearchResource;
import io.swagger.v3.oas.annotations.Hidden;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.headers.Header;
import io.swagger.v3.oas.annotations.media.Content;
import io.swagger.v3.oas.annotations.media.Schema;
import io.swagger.v3.oas.annotations.responses.ApiResponse;
import io.swagger.v3.oas.annotations.responses.ApiResponses;
import lombok.AllArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.opensearch.client.opensearch.core.SearchResponse;
import org.opensearch.client.opensearch.core.search.Hit;
import org.opensearch.client.opensearch.core.search.HitsMetadata;
import org.springdoc.api.annotations.ParameterObject;
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
import java.util.stream.Collectors;

@RestController
@RequestMapping("/api/api-keys")
@PreAuthorize("hasAuthority(\"" + AuthoritiesConstants.USER + "\")")
@Slf4j
@AllArgsConstructor
@Hidden
public class ApiKeyResource {

    private final ApiKeyService apiKeyService;
    private final ElasticsearchService elasticsearchService;
    private final UserService userService;

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
            Long userId = userService.getCurrentUserLogin().getId();
            ApiKeyResponseDTO responseDTO = apiKeyService.createApiKey(userId, dto);
            return ResponseEntity.status(HttpStatus.CREATED).body(responseDTO);
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
    public ResponseEntity<String> generateApiKey(@PathVariable("id") Long apiKeyId) {
        Long userId = userService.getCurrentUserLogin().getId();
        String plainKey = apiKeyService.generateApiKey(userId, apiKeyId);
        return ResponseEntity.ok(plainKey);
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
    public ResponseEntity<ApiKeyResponseDTO> getApiKey(@PathVariable("id") Long apiKeyId) {
        Long userId = userService.getCurrentUserLogin().getId();
        ApiKeyResponseDTO responseDTO = apiKeyService.getApiKey(userId, apiKeyId);
        return ResponseEntity.ok(responseDTO);
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
    @GetMapping
    public ResponseEntity<List<ApiKeyResponseDTO>> listApiKeys(@ParameterObject Pageable pageable) {
        Long userId = userService.getCurrentUserLogin().getId();
        Page<ApiKeyResponseDTO> page = apiKeyService.listApiKeys(userId,pageable);
        HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(ServletUriComponentsBuilder.fromCurrentRequest(), page);

        return ResponseEntity.ok().headers(headers).body(page.getContent());
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
    public ResponseEntity<ApiKeyResponseDTO> updateApiKey(@PathVariable("id") Long apiKeyId,
                                                          @RequestBody ApiKeyUpsertDTO dto) {

        Long userId = userService.getCurrentUserLogin().getId();
        ApiKeyResponseDTO responseDTO = apiKeyService.updateApiKey(userId, apiKeyId, dto);
        return ResponseEntity.ok(responseDTO);

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
    public ResponseEntity<Void> deleteApiKey(@PathVariable("id") Long apiKeyId) {

        Long userId = userService.getCurrentUserLogin().getId();
        apiKeyService.deleteApiKey(userId, apiKeyId);
        return ResponseEntity.noContent().build();
    }

    @PostMapping("/usage")
    public ResponseEntity<List<Map>> search(@RequestBody(required = false) List<FilterType> filters,
                                            @RequestParam Integer top, @RequestParam String indexPattern,
                                            @RequestParam(required = false, defaultValue = "false") boolean includeChildren,
                                            Pageable pageable) {

        SearchResponse<Map> searchResponse = elasticsearchService.search(filters, top, indexPattern,
                pageable, Map.class);

        if (Objects.isNull(searchResponse) || Objects.isNull(searchResponse.hits()) || searchResponse.hits().total().value() == 0)
            return ResponseEntity.ok(Collections.emptyList());

        HitsMetadata<Map> hits = searchResponse.hits();
        HttpHeaders headers = UtilPagination.generatePaginationHttpHeaders(Math.min(hits.total().value(), top),
                pageable.getPageNumber(), pageable.getPageSize(), "/api/elasticsearch/search");

        return ResponseEntity.ok().headers(headers).body(hits.hits().stream()
                .map(Hit::source).collect(Collectors.toList()));

    }
}
