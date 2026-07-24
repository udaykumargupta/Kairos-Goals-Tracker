package com.kairos.auth.dto;

import jakarta.validation.constraints.Email;
import jakarta.validation.constraints.NotBlank;

/** Body of {@code POST /api/auth/forgot-password}: which email to send a reset link to. */
public record ForgotPasswordRequest(@NotBlank @Email String email) {}
