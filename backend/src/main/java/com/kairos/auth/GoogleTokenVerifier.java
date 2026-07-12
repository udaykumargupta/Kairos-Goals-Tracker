package com.kairos.auth;

/**
 * Verifies a Google ID token and returns the identity it asserts.
 *
 * <p>An interface (not a concrete class) so the verification mechanism can be swapped
 * or stubbed in tests without touching {@link AuthService} — Dependency Inversion.
 */
public interface GoogleTokenVerifier {

    /**
     * @param idToken the raw JWT issued by Google Identity Services on the client
     * @return the verified user identity
     * @throws com.kairos.common.AuthException if the token is missing, malformed,
     *                                          expired, or has the wrong audience
     */
    GoogleUser verify(String idToken);
}
