package com.park.utmstack.repository.api_key;

import com.park.utmstack.domain.api_keys.ApiKey;
import org.springframework.cache.annotation.Cacheable;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import javax.validation.constraints.NotNull;
import java.time.Instant;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Repository
public interface ApiKeyRepository extends JpaRepository<ApiKey, Long> {

    Optional<ApiKey> findByIdAndUserId(Long id, Long userId);

    Page<ApiKey> findByUserId(Long userId, Pageable pageable);

    @Cacheable(cacheNames = "apikey", key = "#root.args[0]")
    Optional<ApiKey> findOneByApiKey(@NotNull String apiKey);

    Optional<ApiKey> findByNameAndUserId(@NotNull String name, Long userId);

    List<ApiKey> findAllByExpiresAtAfterAndExpiresAtLessThanEqual(Instant now, Instant fiveDaysFromNow);
}
