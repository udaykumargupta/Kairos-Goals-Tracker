package com.kairos.auth;

import com.kairos.auth.dto.AuthResponse;
import com.kairos.auth.dto.ForgotPasswordRequest;
import com.kairos.auth.dto.GoogleLoginRequest;
import com.kairos.auth.dto.LoginRequest;
import com.kairos.auth.dto.RegisterRequest;
import com.kairos.auth.dto.ResetPasswordRequest;
import jakarta.validation.Valid;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.servlet.support.ServletUriComponentsBuilder;

import java.util.Map;

@RestController
@RequestMapping("/api/auth")
public class AuthController {

    private final AuthService authService;
    private final String configuredBaseUrl;

    public AuthController(AuthService authService,
                          @Value("${kairos.app.base-url:}") String configuredBaseUrl) {
        this.authService = authService;
        this.configuredBaseUrl = configuredBaseUrl;
    }

    /** Exchange a Google ID token for a Kairos JWT. */
    @PostMapping("/google")
    public AuthResponse google(@Valid @RequestBody GoogleLoginRequest request) {
        return authService.loginWithGoogle(request.idToken());
    }

    /** Create a new email/password account and return a Kairos JWT. */
    @PostMapping("/register")
    public AuthResponse register(@Valid @RequestBody RegisterRequest request) {
        return authService.register(request.email(), request.password(), request.name());
    }

    /** Sign in with email + password and return a Kairos JWT. */
    @PostMapping("/login")
    public AuthResponse login(@Valid @RequestBody LoginRequest request) {
        return authService.login(request.email(), request.password());
    }

    /** Email a password-reset link. Always returns the same message (no account enumeration). */
    @PostMapping("/forgot-password")
    public Map<String, String> forgotPassword(@Valid @RequestBody ForgotPasswordRequest request) {
        authService.requestPasswordReset(request.email(), baseUrl());
        return Map.of("message", "If that email is registered, a reset link is on its way.");
    }

    /** Complete a password reset with the emailed token and sign the user in. */
    @PostMapping("/reset-password")
    public AuthResponse resetPassword(@Valid @RequestBody ResetPasswordRequest request) {
        return authService.resetPassword(request.token(), request.newPassword());
    }

    /** Public base URL for links in emails: the configured override, else the request origin. */
    private String baseUrl() {
        if (configuredBaseUrl != null && !configuredBaseUrl.isBlank()) {
            return configuredBaseUrl;
        }
        return ServletUriComponentsBuilder.fromCurrentContextPath().build().toUriString();
    }
}
