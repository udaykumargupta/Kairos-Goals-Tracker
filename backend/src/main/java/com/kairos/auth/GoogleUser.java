package com.kairos.auth;

/**
 * The verified identity extracted from a Google ID token.
 *
 * @param sub           stable, unique Google account id
 * @param email         account email
 * @param emailVerified whether Google has verified the email
 * @param name          display name (may be null)
 * @param picture       avatar URL (may be null)
 */
public record GoogleUser(
        String sub,
        String email,
        boolean emailVerified,
        String name,
        String picture
) {}
