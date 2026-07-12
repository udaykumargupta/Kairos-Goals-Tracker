package com.kairos.share;

import com.kairos.security.AuthenticatedUser;
import com.kairos.share.dto.ShareStatus;
import com.kairos.share.dto.SharedProfile;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class ShareController {

    private final ShareService shareService;

    public ShareController(ShareService shareService) {
        this.shareService = shareService;
    }

    // --- authenticated: manage my own share link ---

    @GetMapping("/api/share/status")
    public ShareStatus status(@AuthenticationPrincipal AuthenticatedUser user) {
        return shareService.status(user.id());
    }

    @PostMapping("/api/share/enable")
    public ShareStatus enable(@AuthenticationPrincipal AuthenticatedUser user) {
        return shareService.enable(user.id());
    }

    @PostMapping("/api/share/disable")
    public ShareStatus disable(@AuthenticationPrincipal AuthenticatedUser user) {
        return shareService.disable(user.id());
    }

    // --- public: view someone's shared progress (read-only) ---

    @GetMapping("/api/public/share/{token}")
    public SharedProfile shared(@PathVariable String token) {
        return shareService.getByToken(token);
    }
}
