package com.park.utmstack.service.api_key;

import com.park.utmstack.domain.api_keys.ApiKey;
import com.park.utmstack.domain.api_keys.ApiKeyUsageLog;
import com.park.utmstack.repository.api_key.ApiKeyRepository;
import com.park.utmstack.service.UserService;
import com.park.utmstack.service.dto.api_key.ApiKeyResponseDTO;
import com.park.utmstack.service.dto.api_key.ApiKeyUpsertDTO;
import com.park.utmstack.service.elasticsearch.OpensearchClientBuilder;
import com.park.utmstack.service.mapper.ApiKeyMapper;
import com.park.utmstack.util.exceptions.ApiKeyExistException;
import com.park.utmstack.util.exceptions.ApiKeyNotFoundException;
import lombok.AllArgsConstructor;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;

import org.springframework.scheduling.annotation.Async;
import org.springframework.stereotype.Service;

import java.security.SecureRandom;
import java.time.Duration;
import java.time.Instant;
import java.util.Base64;
import java.util.Optional;
import java.util.UUID;

import static com.park.utmstack.config.Constants.V11_API_ACCESS_LOGS;

@Service
@AllArgsConstructor
public class ApiKeyService {
    
    private static final String CLASSNAME = "ApiKeyService";
    private final Logger log = LoggerFactory.getLogger(ApiKeyService.class);
    private final ApiKeyRepository apiKeyRepository;
    private final ApiKeyMapper apiKeyMapper;
    private final OpensearchClientBuilder client;


    public ApiKeyResponseDTO createApiKey(Long userId,ApiKeyUpsertDTO dto) {
        final String ctx = CLASSNAME + ".createApiKey";
        try {
            apiKeyRepository.findByNameAndUserId(dto.getName(), userId)
                .ifPresent(apiKey -> {
                    throw new ApiKeyExistException("Api key already exists");
                });

            var apiKey = ApiKey.builder()
                .userId(userId)
                .name(dto.getName())
                .expiresAt(dto.getExpiresAt())
                .allowedIp(String.join(",", dto.getAllowedIp()))
                .createdAt(Instant.now())
                .generatedAt(Instant.now())
                .apiKey(generateRandomKey())
                .build();

            return apiKeyMapper.toDto(apiKeyRepository.save(apiKey));
        } catch (Exception e) {
            throw new ApiKeyExistException(ctx + ": " + e.getMessage());
        }
    }

    public String generateApiKey(Long userId, Long apiKeyId) {
        final String ctx = CLASSNAME + ".generateApiKey";
        try {
            ApiKey apiKey = apiKeyRepository.findByIdAndUserId(apiKeyId, userId)
                .orElseThrow(() -> new ApiKeyNotFoundException("API key not found"));

            Instant now = Instant.now();
            Instant originalCreated = apiKey.getGeneratedAt() != null ? apiKey.getGeneratedAt() : apiKey.getCreatedAt();
            Instant originalExpires = apiKey.getExpiresAt();

            Duration duration;
            if (originalCreated != null && originalExpires != null && !originalExpires.isBefore(originalCreated)) {
                duration = Duration.between(originalCreated, originalExpires);
            } else {
                duration = Duration.ofDays(7);
            }

            String plainKey = generateRandomKey();
            apiKey.setApiKey(plainKey);
            apiKey.setGeneratedAt(Instant.now());
            apiKey.setExpiresAt(now.plus(duration));
            apiKeyRepository.save(apiKey);
            return plainKey;
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    public ApiKeyResponseDTO updateApiKey(Long userId, Long apiKeyId, ApiKeyUpsertDTO dto) {
        final String ctx = CLASSNAME + ".updateApiKey";
        try {
            ApiKey apiKey = apiKeyRepository.findByIdAndUserId(apiKeyId, userId)
                .orElseThrow(() -> new ApiKeyNotFoundException("API key not found"));
            apiKey.setName(dto.getName());
            if (dto.getAllowedIp() != null) {
                apiKey.setAllowedIp(String.join(",", dto.getAllowedIp()));
            } else {
                apiKey.setAllowedIp(null);
            }
            apiKey.setExpiresAt(dto.getExpiresAt());
            ApiKey updated = apiKeyRepository.save(apiKey);
            return apiKeyMapper.toDto(updated);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    public ApiKeyResponseDTO getApiKey(Long userId, Long apiKeyId) {
        final String ctx = CLASSNAME + ".getApiKey";
        try {
            ApiKey apiKey = apiKeyRepository.findByIdAndUserId(apiKeyId, userId)
                .orElseThrow(() -> new ApiKeyNotFoundException("API key not found"));
            return apiKeyMapper.toDto(apiKey);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    public Page<ApiKeyResponseDTO> listApiKeys(Long userId, Pageable pageable) {
        final String ctx = CLASSNAME + ".listApiKeys";
        try {
            return apiKeyRepository.findByUserId(userId, pageable).map(apiKeyMapper::toDto);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }


    public void deleteApiKey(Long userId, Long apiKeyId) {
        final String ctx = CLASSNAME + ".deleteApiKey";
        try {
            ApiKey apiKey = apiKeyRepository.findByIdAndUserId(apiKeyId, userId)
                .orElseThrow(() -> new ApiKeyNotFoundException("API key not found"));
            apiKeyRepository.delete(apiKey);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    private String generateRandomKey() {
        final String ctx = CLASSNAME + ".generateRandomKey";
        try {
            SecureRandom random = new SecureRandom();
            byte[] keyBytes = new byte[32];
            random.nextBytes(keyBytes);
            return Base64.getUrlEncoder().withoutPadding().encodeToString(keyBytes);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    @Async
    public void logUsage(ApiKeyUsageLog apiKeyUsageLog) {
        final String ctx = CLASSNAME + ".logUsage";
        try {
            client.getClient().index(V11_API_ACCESS_LOGS, apiKeyUsageLog);
        } catch (Exception e) {
            log.error(ctx + ": {}", e.getMessage());
        }
    }

    public Optional<ApiKey> findOneByApiKey(String apiKey) {
        return apiKeyRepository.findOneByApiKey(apiKey);
    }


    /*@Scheduled(cron = "0 0 9 * * ?")
    public void checkExpiringApiKeys() {
        Instant fiveDaysFromNow = Instant.now().plus(5, ChronoUnit.DAYS);
        Instant now = Instant.now();
        List<ApiKey> expiringKeys = apiKeyRepository.findAllByExpiresAtAfterAndExpiresAtLessThanEqual(now, fiveDaysFromNow);

        if (!expiringKeys.isEmpty()) {
            Map<Long, List<ApiKey>> expiringKeysByAccount = expiringKeys.stream()
                .collect(Collectors.groupingBy(ApiKey::getUserId));

            expiringKeysByAccount.forEach((userId, apiKeys) -> {
                var principal = userRepository.findByuserIdAndAccountOwnerIsTrue(userId.toString()).orElse(null);
                if (principal == null) {
                    return;
                }
                mailService.sendKeyExpirationEmail(principal, apiKeys);

                userNotificationService.createAndSendNotification(principal.getUuid(),
                    NotificationMessageKeyEnum.apiKey_EXPIRATION,
                    Map.of("names", apiKeys.stream().map(ApiKey::getName).collect(Collectors.joining(","))));
            });
        }
    }*/
}
