package com.park.utmstack.web.rest.user_auditor;

import org.junit.jupiter.api.Test;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;

import java.util.List;

import static org.junit.jupiter.api.Assertions.*;

class UtmAuditorUsersResourceRelayTest {

    @Test
    void relayResponseStripsHopByHopHeadersAndPreservesPaginationAndContentType() {
        HttpHeaders upstreamHeaders = new HttpHeaders();
        upstreamHeaders.add(HttpHeaders.TRANSFER_ENCODING, "chunked");
        upstreamHeaders.add(HttpHeaders.CONNECTION, "keep-alive");
        upstreamHeaders.add("Keep-Alive", "timeout=60");
        upstreamHeaders.add(HttpHeaders.CONTENT_LENGTH, "123");
        upstreamHeaders.setContentType(MediaType.APPLICATION_JSON);
        upstreamHeaders.add("X-Total-Count", "42");
        upstreamHeaders.add(HttpHeaders.LINK, "</api/utm-auditor-users-by-src?page=1>; rel=\"next\"");

        ResponseEntity<List<String>> upstream = ResponseEntity
            .status(HttpStatus.OK)
            .headers(upstreamHeaders)
            .body(List.of("alice", "bob"));

        ResponseEntity<List<String>> relayed = UtmAuditorUsersResource.relayResponse(upstream);

        assertEquals(HttpStatus.OK, relayed.getStatusCode());
        assertEquals(List.of("alice", "bob"), relayed.getBody());
        assertNull(relayed.getHeaders().getFirst(HttpHeaders.TRANSFER_ENCODING));
        assertNull(relayed.getHeaders().getFirst(HttpHeaders.CONNECTION));
        assertNull(relayed.getHeaders().getFirst("Keep-Alive"));
        assertNull(relayed.getHeaders().getFirst(HttpHeaders.CONTENT_LENGTH));
        assertEquals(MediaType.APPLICATION_JSON, relayed.getHeaders().getContentType());
        assertEquals("42", relayed.getHeaders().getFirst("X-Total-Count"));
        assertEquals("</api/utm-auditor-users-by-src?page=1>; rel=\"next\"", relayed.getHeaders().getFirst(HttpHeaders.LINK));
    }
}
