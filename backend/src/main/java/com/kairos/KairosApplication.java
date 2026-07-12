package com.kairos;

import com.kairos.config.AppProperties;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.boot.autoconfigure.security.servlet.UserDetailsServiceAutoConfiguration;
import org.springframework.boot.context.properties.EnableConfigurationProperties;

/**
 * Entry point for the Kairos backend.
 *
 * <p>Responsibilities are intentionally split across small, single-purpose beans
 * (see the {@code auth}, {@code user}, {@code state} and {@code security} packages)
 * so this class only bootstraps Spring.
 */
@SpringBootApplication(exclude = UserDetailsServiceAutoConfiguration.class)
@EnableConfigurationProperties(AppProperties.class)
public class KairosApplication {
    public static void main(String[] args) {
        SpringApplication.run(KairosApplication.class, args);
    }
}
