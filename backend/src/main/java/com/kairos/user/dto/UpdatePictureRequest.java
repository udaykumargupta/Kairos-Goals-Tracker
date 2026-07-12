package com.kairos.user.dto;

/**
 * Body of {@code PUT /api/profile/picture}.
 *
 * @param picture a data URL (e.g. {@code data:image/jpeg;base64,...}); null or blank reverts
 *                to the Google profile picture
 */
public record UpdatePictureRequest(String picture) {}
