package com.park.utmstack.web.rest.idp_provider;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.Map;

@Slf4j
@RestController
@RequestMapping("/api/v1/oauth2")
@RequiredArgsConstructor
public class OAuth2Controller {

    private final DynamicOAuth2Service oauth2Service;

    /**
     * Endpoint para iniciar el flujo OAuth2
     * GET /api/v1/oauth2/authorize/google?redirect_url=http://frontend.com/callback
     */
    @GetMapping("/authorize/{provider}")
    public ResponseEntity<Map<String, String>> authorize(
            @PathVariable String provider,
            @RequestParam(required = false, defaultValue = "/") String redirect_url) {

        try {
            String authorizationUrl = oauth2Service.generateAuthorizationUrl(provider, redirect_url);

            return ResponseEntity.ok(Map.of(
                    "authorization_url", authorizationUrl,
                    "provider", provider
            ));
        } catch (Exception e) {
            log.error("Error generating authorization URL: {}", e.getMessage());
            return ResponseEntity.badRequest()
                    .body(Map.of("error", e.getMessage()));
        }
    }

    /**
     * Lista todos los providers activos
     * GET /api/v1/oauth2/providers
     */
    @GetMapping("/providers")
    public ResponseEntity<?> getActiveProviders() {
        // Implementar según necesidad
        return ResponseEntity.ok(Map.of("message", "List of active providers"));
    }
}
