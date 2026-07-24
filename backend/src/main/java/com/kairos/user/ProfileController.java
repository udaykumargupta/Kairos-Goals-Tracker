package com.kairos.user;

import com.kairos.auth.dto.UserDto;
import com.kairos.security.AuthenticatedUser;
import com.kairos.user.dto.UpdatePictureRequest;
import com.kairos.user.dto.UpdateProfileRequest;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

/**
 * The signed-in user's own profile: read the effective name/email/picture, and set a
 * custom display name.
 */
@RestController
@RequestMapping("/api/profile")
public class ProfileController {

    private final UserService userService;

    public ProfileController(UserService userService) {
        this.userService = userService;
    }

    @GetMapping
    public UserDto get(@AuthenticationPrincipal AuthenticatedUser user) {
        return UserDto.from(userService.getById(user.id()));
    }

    @PutMapping
    public UserDto update(@AuthenticationPrincipal AuthenticatedUser user,
                          @RequestBody UpdateProfileRequest request) {
        return UserDto.from(userService.updateDisplayName(user.id(), request.name()));
    }

    @PutMapping("/picture")
    public UserDto updatePicture(@AuthenticationPrincipal AuthenticatedUser user,
                                 @RequestBody UpdatePictureRequest request) {
        return UserDto.from(userService.updatePicture(user.id(), request.picture()));
    }
}
