package com.kairos.security;

/**
 * The authenticated principal placed in the Spring {@code SecurityContext} for each
 * request. Reconstructed from the verified JWT claims — no database lookup required
 * per request.
 */
public record AuthenticatedUser(Long id, String email, String name, String picture) {}
