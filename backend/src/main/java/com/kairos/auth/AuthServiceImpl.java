package com.kairos.auth;

import com.kairos.auth.dto.AuthResponse;
import com.kairos.auth.dto.UserDto;
import com.kairos.security.JwtService;
import com.kairos.user.User;
import com.kairos.user.UserService;
import org.springframework.stereotype.Service;

/**
 * Composes the collaborators needed for sign-in. Each collaborator has a single
 * responsibility ({@link GoogleTokenVerifier} verifies, {@link UserService} persists,
 * {@link JwtService} mints); this service just wires the steps together.
 */
@Service
public class AuthServiceImpl implements AuthService {

    private final GoogleTokenVerifier googleTokenVerifier;
    private final UserService userService;
    private final JwtService jwtService;

    public AuthServiceImpl(GoogleTokenVerifier googleTokenVerifier,
                           UserService userService,
                           JwtService jwtService) {
        this.googleTokenVerifier = googleTokenVerifier;
        this.userService = userService;
        this.jwtService = jwtService;
    }

    @Override
    public AuthResponse loginWithGoogle(String idToken) {
        GoogleUser googleUser = googleTokenVerifier.verify(idToken);
        User user = userService.upsertFromGoogle(googleUser);
        return issueFor(user);
    }

    @Override
    public AuthResponse register(String email, String password, String displayName) {
        return issueFor(userService.registerLocal(email, password, displayName));
    }

    @Override
    public AuthResponse login(String email, String password) {
        return issueFor(userService.loginLocal(email, password));
    }

    private AuthResponse issueFor(User user) {
        String jwt = jwtService.issue(user);
        return new AuthResponse(jwt, UserDto.from(user));
    }
}
