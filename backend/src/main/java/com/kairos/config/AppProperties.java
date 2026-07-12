package com.kairos.config;

import org.springframework.boot.context.properties.ConfigurationProperties;

/**
 * Strongly-typed application configuration, bound from the {@code kairos.*} keys
 * in {@code application.yml}. Keeping config in one immutable object (rather than
 * scattering {@code @Value} annotations) keeps configuration a single responsibility.
 */
@ConfigurationProperties(prefix = "kairos")
public record AppProperties(Google google, Jwt jwt) {

    public record Google(
            /** OAuth 2.0 Web Client ID from Google Cloud Console; used as the token audience. */
            String clientId
    ) {}

    public record Jwt(
            /** HMAC signing secret. MUST be at least 32 characters (256 bits) for HS256. */
            String secret,
            /** Token lifetime in milliseconds. */
            long expirationMs
    ) {}
}
