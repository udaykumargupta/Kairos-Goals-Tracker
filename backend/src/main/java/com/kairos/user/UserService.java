package com.kairos.user;

import com.kairos.auth.GoogleUser;

import java.util.Optional;

/**
 * User lifecycle operations. Kept as an interface so callers depend on behaviour,
 * not on the JPA-backed implementation.
 */
public interface UserService {

    /**
     * Find the user for this Google identity, creating them on first sign-in and
     * refreshing their cached profile on subsequent sign-ins.
     */
    User upsertFromGoogle(GoogleUser googleUser);

    /** Create a new email/password account. Fails if the email is already registered. */
    User registerLocal(String email, String rawPassword, String displayName);

    /** Authenticate an email/password account, returning the user on success. */
    User loginLocal(String email, String rawPassword);

    /** Find a user by email (case-insensitive), if any. */
    Optional<User> findByEmail(String email);

    /** Set a new password for the user (used by the password-reset flow). */
    User applyNewPassword(Long userId, String rawPassword);

    User getById(Long userId);

    /** Set the user's custom display name; a null/blank value reverts to the Google name. */
    User updateDisplayName(Long userId, String displayName);

    /** Set the user's custom avatar (a data URL); a null/blank value reverts to the Google picture. */
    User updatePicture(Long userId, String pictureDataUrl);
}
