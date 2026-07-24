package com.kairos.auth;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.transaction.annotation.Transactional;

import java.util.Optional;

public interface PasswordResetTokenRepository extends JpaRepository<PasswordResetToken, Long> {

    Optional<PasswordResetToken> findByTokenHash(String tokenHash);

    /** Drop any outstanding tokens for a user (called when a new reset is requested). */
    @Transactional
    void deleteByUserId(Long userId);
}
