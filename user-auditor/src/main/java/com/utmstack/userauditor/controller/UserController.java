package com.utmstack.userauditor.controller;

import com.utmstack.userauditor.controller.utils.UtilPagination;
import com.utmstack.userauditor.controller.utils.UtilResponse;

import com.utmstack.userauditor.model.User;
import com.utmstack.userauditor.service.UserService;
import com.utmstack.userauditor.service.elasticsearch.ElasticsearchService;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

import org.opensearch.client.opensearch.core.SearchResponse;
import org.opensearch.client.opensearch.core.search.Hit;
import org.opensearch.client.opensearch.core.search.HitsMetadata;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.Collections;
import java.util.List;
import java.util.Map;
import java.util.Objects;
import java.util.stream.Collectors;

/**
 * REST controller for managing {@link UtmAuditorUsers}.
 */
@RequiredArgsConstructor
@Slf4j
@RestController
@RequestMapping("/api")
public class UserController {

    private final UserService userService;
    private final ElasticsearchService elasticsearchService;
    private static final String CLASSNAME = "UserResource";

    @GetMapping("/utm-auditor-users-by-src")
    public ResponseEntity<List<User>> getUtmAuditorUsersBySourceId(Pageable pageable, Long id) {
        log.debug("REST request to get a page of UtmAuditorUsers");
        Page<User> page = userService.findBySource(pageable, id);

        return ResponseEntity.ok().body(page.getContent());
    }

    @GetMapping("/search")
    public ResponseEntity<List<Map>> search(@RequestParam String sid,
                                            @RequestParam String from,
                                            @RequestParam String to,
                                            @RequestParam String indexPattern,
                                            @RequestParam String sort,
                                            @RequestParam String page,
                                            @RequestParam String size,
                                            Pageable pageable) {
        final String ctx = CLASSNAME + ".search";
        try {
            SearchResponse<Map> searchResponse = elasticsearchService.searchBySid(sid, to, from, indexPattern, pageable, Map.class);

            if (Objects.isNull(searchResponse) || Objects.isNull(searchResponse.hits()) || searchResponse.hits().total().value() == 0)
                return ResponseEntity.ok(Collections.emptyList());

            HitsMetadata<Map> hits = searchResponse.hits();
            HttpHeaders headers = UtilPagination.generatePaginationHttpHeaders(hits.total().value(),
                    pageable.getPageNumber(), pageable.getPageSize(), "/api/elasticsearch/search");

            return ResponseEntity.ok().headers(headers).body(hits.hits().stream()
                    .map(Hit::source).collect(Collectors.toList()));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }
}
