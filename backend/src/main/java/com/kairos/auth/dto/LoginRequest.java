package com.kairos.auth.dto;

import jakarta.validation.constraints.Email;
import jakarta.validation.constraints.NotBlank;

/** Body of {@code POST /api/auth/login}: authenticate an email/password account. */
public record LoginRequest(
        @NotBlank @Email String email,
        @NotBlank String password
) {}
