package com.utmstack.userauditor.repository;

import com.utmstack.userauditor.model.UserSource;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;

/**
 * Spring Data SQL repository for the UtmAuditorUserSources entity.
 */

@Repository
public interface UserSourceRepository extends JpaRepository<UserSource, Long> {
    List<UserSource> findAllByActiveIsTrue();
}
