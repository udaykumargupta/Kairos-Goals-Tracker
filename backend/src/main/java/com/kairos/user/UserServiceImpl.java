package com.kairos.user;

import com.kairos.auth.GoogleUser;
import com.kairos.common.AuthException;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

@Service
public class UserServiceImpl implements UserService {

    private final UserRepository userRepository;

    public UserServiceImpl(UserRepository userRepository) {
        this.userRepository = userRepository;
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
