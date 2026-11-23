package com.park.utmstack.web.rest.util;

import com.park.utmstack.util.ResponseUtil;
import org.junit.jupiter.api.Test;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;

import static org.junit.jupiter.api.Assertions.*;

class HeaderUtilSanitizationTest {

    @Test
    void failureAlertStripsCrLfAndControlChars() {
        String multiLine = "JDBC error\n  Detail: bad row\r\n  Hint: cast\tvalue";
        HttpHeaders headers = HeaderUtil.createFailureAlert("entity\nname", "key", multiLine);
        String err = headers.getFirst("X-UtmStack-error");
        String params = headers.getFirst("X-UtmStack-params");

        assertNotNull(err);
        assertFalse(err.contains("\n"));
        assertFalse(err.contains("\r"));
        assertFalse(err.contains("\t"));
        assertTrue(err.contains("JDBC error"));
        assertEquals("entity name", params);
    }

    @Test
    void failureAlertWithCodeHandlesNullAndCrLf() {
        HttpHeaders headers = HeaderUtil.createFailureAlertWithCode("line1\r\nline2", null, "ERR\n1");
        assertEquals("line1 line2", headers.getFirst("X-UtmStack-error"));
        assertEquals("", headers.getFirst("X-UtmStack-params"));
        assertEquals("ERR 1", headers.getFirst("X-UtmStack-code"));
    }

    @Test
    void responseUtilBuildErrorResponseHandlesNullMessage() {
        ResponseEntity<Void> resp = ResponseUtil.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, null);
        assertEquals(HttpStatus.INTERNAL_SERVER_ERROR, resp.getStatusCode());
        assertEquals("", resp.getHeaders().getFirst("X-UtmStack-error"));
    }
}
