package com.kairos.common;

/**
 * Thrown when a requested resource does not exist (e.g. an unknown/disabled share link).
 * Mapped to HTTP 404 by {@link GlobalExceptionHandler}.
 */
public class NotFoundException extends RuntimeException {

    public NotFoundException(String message) {
        super(message);
    }
}
