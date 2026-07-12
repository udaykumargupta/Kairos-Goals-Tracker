package com.kairos.common;

/**
 * Thrown when authentication fails (e.g. an invalid Google ID token).
 * Mapped to HTTP 401 by {@link GlobalExceptionHandler}.
 */
public class AuthException extends RuntimeException {

    public AuthException(String message) {
        super(message);
    }

    public AuthException(String message, Throwable cause) {
        super(message, cause);
    }
}
