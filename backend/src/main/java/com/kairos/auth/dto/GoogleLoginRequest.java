package com.kairos.auth.dto;

import jakarta.validation.constraints.NotBlank;

/** Body of {@code POST /api/auth/google}: the Google ID token obtained on the client. */
public record GoogleLoginRequest(@NotBlank String idToken) {}
