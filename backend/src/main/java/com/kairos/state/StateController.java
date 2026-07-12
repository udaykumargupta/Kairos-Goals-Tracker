package com.kairos.state;

import com.fasterxml.jackson.databind.JsonNode;
import com.kairos.security.AuthenticatedUser;
import com.kairos.state.dto.StateResponse;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

/**
 * The authenticated goal-sync API. The current user is taken from the JWT-derived
 * principal, so a user can only ever read or write their own state.
 */
@RestController
@RequestMapping("/api/state")
public class StateController {

    private final UserStateService stateService;

    public StateController(UserStateService stateService) {
        this.stateService = stateService;
    }

    @GetMapping
    public StateResponse get(@AuthenticationPrincipal AuthenticatedUser user) {
        return stateService.getForUser(user.id());
    }

    @PutMapping
    public StateResponse put(@AuthenticationPrincipal AuthenticatedUser user,
                             @RequestBody JsonNode data) {
        return stateService.saveForUser(user.id(), data);
    }
}
