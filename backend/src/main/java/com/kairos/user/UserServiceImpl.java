package com.kairos.user;

import com.kairos.auth.GoogleUser;
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
}
