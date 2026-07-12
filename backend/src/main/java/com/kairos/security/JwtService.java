package com.kairos.security;

import com.kairos.user.User;

import java.util.Optional;

/**
 * Issues and verifies the application's own JWTs (distinct from Google's ID token).
 * After a successful Google sign-in we mint one of these; the browser then presents it
 * as a bearer token on every API call.
 */
public interface JwtService {

    /** Mint a signed JWT for the given user. */
    String issue(User user);

    /**
     * Verify a token's signature and expiry.
     *
     * @return the authenticated principal, or empty if the token is invalid/expired
     */
    Optional<AuthenticatedUser> verify(String token);
}
