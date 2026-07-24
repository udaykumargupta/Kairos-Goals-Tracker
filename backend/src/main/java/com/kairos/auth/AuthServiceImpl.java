package com.kairos.auth;

import com.kairos.auth.dto.AuthResponse;
import com.kairos.auth.dto.UserDto;
import com.kairos.common.AuthException;
import com.kairos.security.JwtService;
import com.kairos.user.User;
import com.kairos.user.UserService;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.security.SecureRandom;
import java.time.Instant;
import java.time.temporal.ChronoUnit;
import java.util.Base64;
import java.util.HexFormat;
import java.util.Optional;

/**
 * Composes the collaborators needed for sign-in. Each collaborator has a single
 * responsibility ({@link GoogleTokenVerifier} verifies, {@link UserService} persists,
 * {@link JwtService} mints); this service just wires the steps together.
 */
@Service
public class AuthServiceImpl implements AuthService {

    private static final long RESET_TOKEN_TTL_MINUTES = 60;
    private static final SecureRandom RANDOM = new SecureRandom();

    private final GoogleTokenVerifier googleTokenVerifier;
    private final UserService userService;
    private final JwtService jwtService;
    private final PasswordResetTokenRepository resetTokenRepository;
    private final EmailService emailService;

    public AuthServiceImpl(GoogleTokenVerifier googleTokenVerifier,
                           UserService userService,
                           JwtService jwtService,
                           PasswordResetTokenRepository resetTokenRepository,
                           EmailService emailService) {
        this.googleTokenVerifier = googleTokenVerifier;
        this.userService = userService;
        this.jwtService = jwtService;
        this.resetTokenRepository = resetTokenRepository;
        this.emailService = emailService;
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

    @Override
    @Transactional
    public void requestPasswordReset(String email, String baseUrl) {
        Optional<User> maybeUser = userService.findByEmail(email);
        if (maybeUser.isEmpty()) {
            return;                                  // silent: don't reveal whether the email exists
        }
        User user = maybeUser.get();
        resetTokenRepository.deleteByUserId(user.getId());   // invalidate previous links

        String rawToken = randomToken();
        Instant expiresAt = Instant.now().plus(RESET_TOKEN_TTL_MINUTES, ChronoUnit.MINUTES);
        resetTokenRepository.save(new PasswordResetToken(user.getId(), sha256(rawToken), expiresAt));

        String base = (baseUrl == null || baseUrl.isBlank()) ? "" : baseUrl.replaceAll("/+$", "");
        String link = base + "/?reset=" + rawToken;
        emailService.sendPasswordReset(user.getEmail(), link);
    }

    @Override
    @Transactional
    public AuthResponse resetPassword(String token, String newPassword) {
        if (token == null || token.isBlank()) {
            throw new AuthException("This reset link is invalid or has expired.");
        }
        PasswordResetToken prt = resetTokenRepository.findByTokenHash(sha256(token))
                .filter(PasswordResetToken::isUsable)
                .orElseThrow(() -> new AuthException("This reset link is invalid or has expired."));
        prt.markUsed();
        resetTokenRepository.save(prt);
        User user = userService.applyNewPassword(prt.getUserId(), newPassword);
        return issueFor(user);
    }

    private AuthResponse issueFor(User user) {
        String jwt = jwtService.issue(user);
        return new AuthResponse(jwt, UserDto.from(user));
    }

    private static String randomToken() {
        byte[] bytes = new byte[32];
        RANDOM.nextBytes(bytes);
        return Base64.getUrlEncoder().withoutPadding().encodeToString(bytes);
    }

    private static String sha256(String value) {
        try {
            MessageDigest md = MessageDigest.getInstance("SHA-256");
            return HexFormat.of().formatHex(md.digest(value.getBytes(StandardCharsets.UTF_8)));
        } catch (Exception e) {
            throw new IllegalStateException("SHA-256 unavailable", e);
        }
    }
}
