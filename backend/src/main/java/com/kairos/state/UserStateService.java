package com.kairos.state;

import com.fasterxml.jackson.databind.JsonNode;
import com.kairos.state.dto.StateResponse;

/**
 * Loads and saves a user's goal document. The controller depends on this interface,
 * not the JPA-backed implementation.
 */
public interface UserStateService {

    /** @return the stored state for the user, or a response with null data if none exists. */
    StateResponse getForUser(Long userId);

    /**
     * Replace the user's stored state with {@code data}.
     *
     * @return the saved state including the new timestamp
     */
    StateResponse saveForUser(Long userId, JsonNode data);
}
