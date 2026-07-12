package com.kairos.share;

import com.kairos.common.AuthException;
import com.kairos.common.NotFoundException;
import com.kairos.share.dto.ShareStatus;
import com.kairos.share.dto.SharedProfile;
import com.kairos.state.UserStateService;
import com.kairos.state.dto.StateResponse;
import com.kairos.user.User;
import com.kairos.user.UserRepository;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.security.SecureRandom;
import java.util.Base64;

@Service
public class ShareServiceImpl implements ShareService {

    private final UserRepository userRepository;
    private final UserStateService stateService;
    private final SecureRandom random = new SecureRandom();

    public ShareServiceImpl(UserRepository userRepository, UserStateService stateService) {
        this.userRepository = userRepository;
        this.stateService = stateService;
    }

    @Override
    @Transactional
    public ShareStatus enable(Long userId) {
        User user = requireUser(userId);
        if (user.getShareToken() == null) {
            user.setShareToken(generateToken());
            userRepository.save(user);
        }
        return new ShareStatus(true, user.getShareToken());
    }

    @Override
    @Transactional
    public ShareStatus disable(Long userId) {
        User user = requireUser(userId);
        user.setShareToken(null);
        userRepository.save(user);
        return new ShareStatus(false, null);
    }

    @Override
    @Transactional(readOnly = true)
    public ShareStatus status(Long userId) {
        User user = requireUser(userId);
        return new ShareStatus(user.getShareToken() != null, user.getShareToken());
    }

    @Override
    @Transactional(readOnly = true)
    public SharedProfile getByToken(String token) {
        User user = userRepository.findByShareToken(token)
                .orElseThrow(() -> new NotFoundException("Share link not found or disabled"));
        StateResponse state = stateService.getForUser(user.getId());
        return new SharedProfile(user.effectiveName(), user.effectivePicture(), state.updatedAt(), state.data());
    }

    private User requireUser(Long userId) {
        return userRepository.findById(userId)
                .orElseThrow(() -> new AuthException("User not found"));
    }

    private String generateToken() {
        byte[] bytes = new byte[16];
        random.nextBytes(bytes);
        return Base64.getUrlEncoder().withoutPadding().encodeToString(bytes);
    }
}
