package com.kairos.user;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.GenerationType;
import jakarta.persistence.Id;
import jakarta.persistence.PrePersist;
import jakarta.persistence.PreUpdate;
import jakarta.persistence.Table;

import java.time.Instant;

/**
 * A Kairos user, identified by their stable Google subject id ({@code sub}).
 * Email/name/picture are cached copies of the Google profile, refreshed on each login.
 */
@Entity
@Table(name = "users")
public class User {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(name = "google_sub", nullable = false, unique = true)
    private String googleSub;

    @Column(nullable = false)
    private String email;

    /** Name from the Google profile, refreshed on each login. */
    private String name;

    /** User-chosen display name; when set, overrides {@link #name}. Not touched by login. */
    @Column(name = "display_name")
    private String displayName;

    /** Avatar URL from the Google profile, refreshed on each login. */
    @Column(length = 1024)
    private String picture;

    /** User-uploaded avatar (a data URL); when set, overrides {@link #picture}. Not touched by login. */
    @Column(name = "custom_picture", columnDefinition = "text")
    private String customPicture;

    /** BCrypt hash for email/password accounts; null for Google-only accounts. */
    @Column(name = "password_hash", columnDefinition = "text")
    private String passwordHash;

    /** When non-null, progress is shared read-only at /?share=&lt;shareToken&gt;. */
    @Column(name = "share_token", unique = true)
    private String shareToken;

    @Column(name = "created_at", nullable = false, updatable = false)
    private Instant createdAt;

    @Column(name = "updated_at", nullable = false)
    private Instant updatedAt;

    protected User() {
        // for JPA
    }

    public User(String googleSub, String email, String name, String picture) {
        this.googleSub = googleSub;
        this.email = email;
        this.name = name;
        this.picture = picture;
    }

    /**
     * Create an email/password user (no Google identity). A synthetic, unique
     * {@code google_sub} keeps the existing NOT NULL/UNIQUE column satisfied
     * without a schema change.
     */
    public static User localUser(String email, String name, String passwordHash) {
        User u = new User();
        u.googleSub = "local:" + java.util.UUID.randomUUID();
        u.email = email;
        u.name = name;
        u.picture = null;
        u.passwordHash = passwordHash;
        return u;
    }

    @PrePersist
    void onCreate() {
        Instant now = Instant.now();
        this.createdAt = now;
        this.updatedAt = now;
    }

    @PreUpdate
    void onUpdate() {
        this.updatedAt = Instant.now();
    }

    /** Refresh the cached Google profile fields. */
    public void updateProfile(String email, String name, String picture) {
        this.email = email;
        this.name = name;
        this.picture = picture;
    }

    public Long getId() {
        return id;
    }

    public String getGoogleSub() {
        return googleSub;
    }

    public String getEmail() {
        return email;
    }

    public String getName() {
        return name;
    }

    public String getDisplayName() {
        return displayName;
    }

    public void setDisplayName(String displayName) {
        this.displayName = displayName;
    }

    /** The name to show: the user's chosen display name if set, else the Google name. */
    public String effectiveName() {
        return (displayName != null && !displayName.isBlank()) ? displayName : name;
    }

    public String getCustomPicture() {
        return customPicture;
    }

    public void setCustomPicture(String customPicture) {
        this.customPicture = customPicture;
    }

    /** The avatar to show: the uploaded picture if set, else the Google picture. */
    public String effectivePicture() {
        return (customPicture != null && !customPicture.isBlank()) ? customPicture : picture;
    }

    public String getPicture() {
        return picture;
    }

    public String getPasswordHash() {
        return passwordHash;
    }

    public void setPasswordHash(String passwordHash) {
        this.passwordHash = passwordHash;
    }

    /** Whether this account can sign in with a password (vs. Google-only). */
    public boolean hasPassword() {
        return passwordHash != null && !passwordHash.isBlank();
    }

    public String getShareToken() {
        return shareToken;
    }

    public void setShareToken(String shareToken) {
        this.shareToken = shareToken;
    }

    public Instant getCreatedAt() {
        return createdAt;
    }

    public Instant getUpdatedAt() {
        return updatedAt;
    }
}
