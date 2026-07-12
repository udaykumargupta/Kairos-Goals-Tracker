package com.kairos.auth.dto;

/** Response of {@code POST /api/auth/google}: the app JWT plus the user's profile. */
public record AuthResponse(String token, UserDto user) {}
