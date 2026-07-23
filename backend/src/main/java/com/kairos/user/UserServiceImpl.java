package com.kairos.user;

import com.kairos.auth.GoogleUser;
import com.kairos.common.AuthException;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

@Service
public class UserServiceImpl implements UserService {

    private final UserRepository userRepository;
    private final PasswordEncoder passwordEncoder;

    public UserServiceImpl(UserRepository userRepository, PasswordEncoder passwordEncoder) {
        this.userRepository = userRepository;
        this.passwordEncoder = passwordEncoder;
    }

    @Override
    @Transactional
    public User upsertFromGoogle(GoogleUser googleUser) {
        return userRepository.findByGoogleSub(googleUser.sub())
                .map(existing -> {
                    existing.updateProfile(googleUser.email(), googleUser.name(), googleUser.picture());
                    return existing; // dirty-checked, flushed on commit
                })
                .orElseGet(() -> userRepository.save(new User(
                        googleUser.sub(),
                        googleUser.email(),
                        googleUser.name(),
                        googleUser.picture()
                )));
    }

    @Override
    @Transactional
    public User registerLocal(String email, String rawPassword, String displayName) {
        String normalized = normalizeEmail(email);
        if (userRepository.existsByEmailIgnoreCase(normalized)) {
            throw new IllegalArgumentException("That email is already registered. Try logging in instead.");
        }
        String name = (displayName != null && !displayName.isBlank())
                ? displayName.trim()
                : normalized.substring(0, normalized.indexOf('@') > 0 ? normalized.indexOf('@') : normalized.length());
        String hash = passwordEncoder.encode(rawPassword);
        return userRepository.save(User.localUser(normalized, name, hash));
    }

    @Override
    @Transactional(readOnly = true)
    public User loginLocal(String email, String rawPassword) {
        User user = userRepository.findFirstByEmailIgnoreCaseOrderByIdAsc(normalizeEmail(email))
                .orElseThrow(() -> new AuthException("Invalid email or password"));
        if (!user.hasPassword()) {
            throw new AuthException("This email is registered with Google — use “Sign in with Google”.");
        }
        if (!passwordEncoder.matches(rawPassword, user.getPasswordHash())) {
            throw new AuthException("Invalid email or password");
        }
        return user;
    }

    @Override
    @Transactional
    public User setPassword(Long userId, String currentPassword, String newPassword) {
        if (newPassword == null || newPassword.length() < 8 || newPassword.length() > 100) {
            throw new IllegalArgumentException("Password must be 8–100 characters");
        }
        User user = getById(userId);
        if (user.hasPassword()) {
            // Bad-request (not 401): the session is valid, only the supplied current password is wrong.
            if (currentPassword == null || !passwordEncoder.matches(currentPassword, user.getPasswordHash())) {
                throw new IllegalArgumentException("Current password is incorrect");
            }
        }
        user.setPasswordHash(passwordEncoder.encode(newPassword));
        return userRepository.save(user);
    }

    private static String normalizeEmail(String email) {
        return email == null ? "" : email.trim().toLowerCase();
    }

    @Override
    @Transactional(readOnly = true)
    public User getById(Long userId) {
        return userRepository.findById(userId)
                .orElseThrow(() -> new AuthException("User not found"));
    }

    @Override
    @Transactional
    public User updateDisplayName(Long userId, String displayName) {
        User user = getById(userId);
        user.setDisplayName(displayName == null || displayName.isBlank() ? null : displayName.trim());
        return userRepository.save(user);
    }

    private static final int MAX_PICTURE_CHARS = 700_000;   // ~500 KB data URL

    @Override
    @Transactional
    public User updatePicture(Long userId, String pictureDataUrl) {
        if (pictureDataUrl != null && pictureDataUrl.length() > MAX_PICTURE_CHARS) {
            throw new IllegalArgumentException("Image is too large; please choose a smaller picture");
        }
        User user = getById(userId);
        user.setCustomPicture(pictureDataUrl == null || pictureDataUrl.isBlank() ? null : pictureDataUrl);
        return userRepository.save(user);
    }
}
