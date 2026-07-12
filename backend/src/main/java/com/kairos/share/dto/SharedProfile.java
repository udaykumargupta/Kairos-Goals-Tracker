package com.kairos.share.dto;

import com.fasterxml.jackson.databind.JsonNode;

/**
 * The read-only snapshot returned for a valid share link.
 *
 * @param name      owner's display name
 * @param picture   owner's avatar URL
 * @param updatedAt epoch millis of the owner's last save (null if never)
 * @param data      the owner's goal + journal document
 */
public record SharedProfile(String name, String picture, Long updatedAt, JsonNode data) {}
