package com.kairos.auth.dto;

import com.kairos.user.User;

/** Public view of a user returned to the client after login. */
public record UserDto(String email, String name, String picture) {

    public static UserDto from(User user) {
        return new UserDto(user.getEmail(), user.effectiveName(), user.effectivePicture());
    }
}
