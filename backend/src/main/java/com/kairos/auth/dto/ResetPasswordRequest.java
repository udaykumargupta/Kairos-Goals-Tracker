package com.kairos.auth.dto;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Size;

/** Body of {@code POST /api/auth/reset-password}: the emailed token plus the new password. */
public record ResetPasswordRequest(
        @NotBlank String token,
        @NotBlank @Size(min = 8, max = 100, message = "must be at least 8 characters") String newPassword
) {}
