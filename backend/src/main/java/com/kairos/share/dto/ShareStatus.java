package com.kairos.share.dto;

/**
 * Current sharing state for the signed-in user.
 *
 * @param enabled whether a public share link is active
 * @param token   the share token (null when disabled)
 */
public record ShareStatus(boolean enabled, String token) {}
