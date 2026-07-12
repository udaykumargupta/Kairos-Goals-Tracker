package com.kairos.auth;

import com.google.api.client.googleapis.auth.oauth2.GoogleIdToken;
import com.google.api.client.googleapis.auth.oauth2.GoogleIdTokenVerifier;
import com.google.api.client.http.javanet.NetHttpTransport;
import com.google.api.client.json.gson.GsonFactory;
import com.kairos.common.AuthException;
import com.kairos.config.AppProperties;
import org.springframework.stereotype.Component;
import org.springframework.util.StringUtils;

import java.io.IOException;
import java.security.GeneralSecurityException;
import java.util.Collections;

/**
 * Verifies Google ID tokens using Google's official client library. The token's
 * signature, expiry, issuer, and — critically — its {@code aud} (audience) claim are
 * all checked against our configured OAuth Client ID.
 */
@Component
public class GoogleTokenVerifierImpl implements GoogleTokenVerifier {

    private final GoogleIdTokenVerifier verifier;

    public GoogleTokenVerifierImpl(AppProperties properties) {
        String clientId = properties.google().clientId();
        this.verifier = new GoogleIdTokenVerifier.Builder(new NetHttpTransport(), GsonFactory.getDefaultInstance())
                .setAudience(Collections.singletonList(clientId))
                .build();
    }

    @Override
    public GoogleUser verify(String idToken) {
        if (!StringUtils.hasText(idToken)) {
            throw new AuthException("Missing Google ID token");
        }
        final GoogleIdToken token;
        try {
            token = verifier.verify(idToken);
        } catch (GeneralSecurityException | IOException | IllegalArgumentException e) {
            // IllegalArgumentException covers malformed (non-JWT) tokens from the parser.
            throw new AuthException("Could not verify Google ID token", e);
        }
        if (token == null) {
            throw new AuthException("Invalid Google ID token");
        }
        GoogleIdToken.Payload payload = token.getPayload();
        Boolean emailVerified = payload.getEmailVerified();
        if (emailVerified == null || !emailVerified) {
            throw new AuthException("Google account email is not verified");
        }
        return new GoogleUser(
                payload.getSubject(),
                payload.getEmail(),
                true,
                (String) payload.get("name"),
                (String) payload.get("picture")
        );
    }
}
