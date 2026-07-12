package com.kairos.share;

import com.kairos.share.dto.ShareStatus;
import com.kairos.share.dto.SharedProfile;

/**
 * Manages read-only progress sharing: enabling/disabling a per-user share token
 * and resolving a token to a public snapshot.
 */
public interface ShareService {

    ShareStatus enable(Long userId);

    ShareStatus disable(Long userId);

    ShareStatus status(Long userId);

    /**
     * @throws com.kairos.common.NotFoundException if the token is unknown or sharing is disabled
     */
    SharedProfile getByToken(String token);
}
