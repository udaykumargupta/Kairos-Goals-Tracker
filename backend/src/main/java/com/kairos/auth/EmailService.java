package com.kairos.auth;

/** Sends transactional emails. Abstracted so the delivery mechanism can be swapped. */
public interface EmailService {

    /** Email a password-reset link to the given address. */
    void sendPasswordReset(String toEmail, String resetLink);
}
