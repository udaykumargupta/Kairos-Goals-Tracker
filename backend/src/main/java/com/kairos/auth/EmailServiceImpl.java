package com.kairos.auth;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.ObjectProvider;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.mail.SimpleMailMessage;
import org.springframework.mail.javamail.JavaMailSender;
import org.springframework.stereotype.Service;

/**
 * Sends email via SMTP when configured (Gmail etc.). If SMTP isn't set up
 * ({@code MAIL_HOST}/{@code MAIL_USERNAME} blank), it falls back to logging the link,
 * so the reset flow is fully testable in dev without an email provider.
 */
@Service
public class EmailServiceImpl implements EmailService {

    private static final Logger log = LoggerFactory.getLogger(EmailServiceImpl.class);

    private final ObjectProvider<JavaMailSender> mailSenderProvider;
    private final String username;
    private final String from;

    public EmailServiceImpl(ObjectProvider<JavaMailSender> mailSenderProvider,
                            @Value("${spring.mail.username:}") String username,
                            @Value("${kairos.mail.from:}") String fromEmail,
                            @Value("${kairos.mail.from-name:Kairos}") String fromName) {
        this.mailSenderProvider = mailSenderProvider;
        this.username = username;
        String email = (fromEmail == null || fromEmail.isBlank()) ? username : fromEmail;
        // Display name is independent of the Gmail username — recipients see "Kairos <email>".
        this.from = (email == null || email.isBlank()) ? email
                : (fromName == null || fromName.isBlank() ? email : fromName + " <" + email + ">");
    }

    @Override
    public void sendPasswordReset(String toEmail, String resetLink) {
        JavaMailSender sender = mailSenderProvider.getIfAvailable();
        boolean configured = sender != null && username != null && !username.isBlank();
        if (!configured) {
            // Dev fallback: no SMTP configured — log the link instead of sending.
            log.info("[DEV] Password reset for {} (email not configured) → {}", toEmail, resetLink);
            return;
        }
        SimpleMailMessage msg = new SimpleMailMessage();
        msg.setFrom(from);
        msg.setTo(toEmail);
        msg.setSubject("Reset your Kairos password");
        msg.setText(
                "Hi,\n\n" +
                "We received a request to set a new password for your Kairos account.\n" +
                "Click the link below to choose a new password (it expires in 1 hour):\n\n" +
                resetLink + "\n\n" +
                "If you didn't request this, you can safely ignore this email.\n\n" +
                "— Kairos"
        );
        try {
            sender.send(msg);
            log.info("Password reset email sent to {}", toEmail);
        } catch (Exception e) {
            log.warn("Failed to send password reset email to {}: {}", toEmail, e.getMessage());
        }
    }
}
