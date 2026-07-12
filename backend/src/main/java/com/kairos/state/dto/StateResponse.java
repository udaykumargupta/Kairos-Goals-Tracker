package com.kairos.state.dto;

import com.fasterxml.jackson.databind.JsonNode;

/**
 * The user's stored goal document plus a last-modified timestamp.
 *
 * @param data      the goal JSON (null if the user has never saved anything)
 * @param updatedAt epoch millis of the last save (null if never saved)
 */
public record StateResponse(JsonNode data, Long updatedAt) {}
