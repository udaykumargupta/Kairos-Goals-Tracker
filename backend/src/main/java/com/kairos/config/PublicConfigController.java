package com.kairos.config;

import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.Map;

/**
 * Exposes non-secret configuration the browser needs — currently just the Google
 * OAuth Client ID (which is public by design). Keeping it server-driven means the
 * Client ID is configured in exactly one place.
 */
@RestController
@RequestMapping("/api/public")
public class PublicConfigController {

    private final AppProperties properties;

    public PublicConfigController(AppProperties properties) {
        this.properties = properties;
    }

    @GetMapping("/config")
    public Map<String, String> config() {
        return Map.of("googleClientId", properties.google().clientId());
    }
}
