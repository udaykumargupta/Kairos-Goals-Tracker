package com.kairos.user;

import org.springframework.data.jpa.repository.JpaRepository;

import java.util.Optional;

/**
 * Persistence boundary for {@link User}. Depending on this Spring Data interface
 * (an abstraction) rather than a concrete DB class keeps the service layer decoupled
 * from JPA specifics (Dependency Inversion).
 */
public interface UserRepository extends JpaRepository<User, Long> {
    Optional<User> findByGoogleSub(String googleSub);

    Optional<User> findByShareToken(String shareToken);

    boolean existsByEmailIgnoreCase(String email);

    Optional<User> findFirstByEmailIgnoreCaseOrderByIdAsc(String email);
}
