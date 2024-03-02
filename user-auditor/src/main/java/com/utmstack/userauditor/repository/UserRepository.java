package com.utmstack.userauditor.repository;

import com.utmstack.userauditor.model.User;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.domain.Sort;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.Optional;

/**
 * Spring Data SQL repository for the UtmAuditorUsers entity.
 */

@Repository
public interface UserRepository extends JpaRepository<User, Long> {
    Page<User> findAllBySourceId(Pageable pageable, @Param("sourceId") Long sourceId);
    Optional<User> findByName(String username);
}
