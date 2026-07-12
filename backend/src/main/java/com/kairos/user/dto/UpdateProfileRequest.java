package com.kairos.user.dto;

/**
 * Body of {@code PUT /api/profile}.
 *
 * @param name the new display name; null or blank reverts to the Google profile name
 */
public record UpdateProfileRequest(String name) {}
