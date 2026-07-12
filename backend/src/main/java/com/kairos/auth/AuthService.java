package com.kairos.auth;

import com.kairos.auth.dto.AuthResponse;

/**
 * Orchestrates sign-in: verify the Google token, resolve/create the user, and issue
 * an application JWT. The controller depends on this abstraction only.
 */
public interface AuthService {

    AuthResponse loginWithGoogle(String idToken);
}
