package com.kairos.user;

import com.kairos.auth.GoogleUser;

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
}
