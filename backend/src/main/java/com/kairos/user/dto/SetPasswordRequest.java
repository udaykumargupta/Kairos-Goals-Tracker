package com.kairos.user.dto;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Size;

/**
 * Body of {@code PUT /api/profile/password}. {@code currentPassword} is required only
 * when the account already has a password (changing it); it's ignored on first-time set.
 */
public record SetPasswordRequest(
        String currentPassword,
        @NotBlank @Size(min = 8, max = 100, message = "must be at least 8 characters") String newPassword
) {}
