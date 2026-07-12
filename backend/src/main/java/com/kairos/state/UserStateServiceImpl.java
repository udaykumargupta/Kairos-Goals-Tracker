package com.kairos.state;

import com.fasterxml.jackson.databind.JsonNode;
import com.kairos.state.dto.StateResponse;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.Instant;

@Service
public class UserStateServiceImpl implements UserStateService {

    private final UserStateRepository repository;

    public UserStateServiceImpl(UserStateRepository repository) {
        this.repository = repository;
    }

    @Override
    @Transactional(readOnly = true)
    public StateResponse getForUser(Long userId) {
        return repository.findById(userId)
                .map(state -> new StateResponse(state.getData(), state.getUpdatedAt().toEpochMilli()))
                .orElse(new StateResponse(null, null));
    }

    @Override
    @Transactional
    public StateResponse saveForUser(Long userId, JsonNode data) {
        if (data == null || data.isNull()) {
            throw new IllegalArgumentException("State data must not be null");
        }
        Instant now = Instant.now();
        UserState state = repository.findById(userId)
                .map(existing -> {
                    existing.update(data, now);
                    return existing;
                })
                .orElseGet(() -> new UserState(userId, data, now));
        repository.save(state);
        return new StateResponse(data, now.toEpochMilli());
    }
}
