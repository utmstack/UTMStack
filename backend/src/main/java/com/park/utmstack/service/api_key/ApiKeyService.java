package com.utmstack.api.service;

import com.utmstack.api.domain.User;
import com.utmstack.api.domain.api_key.ApiKey;
import com.utmstack.api.domain.api_key.ApiKeyUsageLog_;
import com.utmstack.api.domain.api_key.index.ApiKeyUsageLogIndexDocument;
import com.utmstack.api.domain.enumeration.NotificationMessageKeyEnum;
import com.utmstack.api.repository.elasticsearch.ApiKeyUsageLogRepository;
import com.utmstack.api.repository.jpa.ApiKeyRepository;
import com.utmstack.api.repository.jpa.UserRepository;
import com.utmstack.api.service.criteria.api_key.ApiKeyUsageLogCriteria;
import com.utmstack.api.service.dto.SearchHitsResponseDTO;
import com.utmstack.api.service.dto.api_key.ApiKeyResponseDTO;
import com.utmstack.api.service.dto.api_key.ApiKeyUpsertDTO;
import com.utmstack.api.service.exceptions.ApiKeyExistException;
import com.utmstack.api.service.exceptions.ApiKeyNotFoundException;
import com.utmstack.api.service.mapper.ApiKeyMapper;
import com.utmstack.api.service.user.UserNotificationService;
import lombok.AllArgsConstructor;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.elasticsearch.core.ElasticsearchOperations;
import org.springframework.data.elasticsearch.core.SearchHit;
import org.springframework.data.elasticsearch.core.SearchHits;
import org.springframework.data.elasticsearch.core.query.Criteria;
import org.springframework.data.elasticsearch.core.query.CriteriaQuery;
import org.springframework.scheduling.annotation.Async;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;

import java.security.SecureRandom;
import java.time.Instant;
import java.time.temporal.ChronoUnit;
import java.util.Base64;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.UUID;
import java.util.stream.Collectors;

@Service
@AllArgsConstructor
public class ApiKeyService {
    private static final String CLASSNAME = "ApiKeyService";
    private final Logger log = LoggerFactory.getLogger(ApiKeyService.class);
    private final ApiKeyRepository apiKeyRepository;
    private final ApiKeyMapper apiKeyMapper;
    private final ApiKeyUsageLogRepository apiUsageLogRepository;
    private final ElasticsearchOperations elasticsearchOperations;
    private final UserRepository userRepository;
    private final UserNotificationService userNotificationService;
    private final MailService mailService;


    public ApiKeyResponseDTO createApiKey(UUID accountId, ApiKeyUpsertDTO dto) {
        final String ctx = CLASSNAME + ".createApiKey";
        try {
            apiKeyRepository.findByNameAndAccountId(dto.getName(), accountId)
                .ifPresent(apiKey -> {
                    throw new ApiKeyExistException("Api key already exists");
                });
            var apiKey = api_key.builder()
                .accountId(accountId)
                .name(dto.getName())
                .expiresAt(dto.getExpiresAt())
                .allowedIp(String.join(",", dto.getAllowedIp()))
                .createdAt(Instant.now())
                .apiKey(generateRandomKey())
                .build();
            return apiKeyMapper.toDto(apiKeyRepository.save(apiKey));
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    public String generateApiKey(UUID accountId, UUID apiKeyId) {
        final String ctx = CLASSNAME + ".generateApiKey";
        try {
            ApiKey apiKey = apiKeyRepository.findByIdAndAccountId(apiKeyId, accountId)
                .orElseThrow(() -> new ApiKeyNotFoundException("API key not found"));
            String plainKey = generateRandomKey();
            api_key.setApiKey(plainKey);
            api_key.setGeneratedAt(Instant.now());
            apiKeyRepository.save(apiKey);
            return plainKey;
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    public ApiKeyResponseDTO updateApiKey(UUID accountId, UUID apiKeyId, ApiKeyUpsertDTO dto) {
        final String ctx = CLASSNAME + ".updateApiKey";
        try {
            ApiKey apiKey = apiKeyRepository.findByIdAndAccountId(apiKeyId, accountId)
                .orElseThrow(() -> new ApiKeyNotFoundException("API key not found"));
            api_key.setName(dto.getName());
            if (dto.getAllowedIp() != null) {
                api_key.setAllowedIp(String.join(",", dto.getAllowedIp()));
            } else {
                api_key.setAllowedIp(null);
            }
            api_key.setExpiresAt(dto.getExpiresAt());
            ApiKey updated = apiKeyRepository.save(apiKey);
            return apiKeyMapper.toDto(updated);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    public ApiKeyResponseDTO getApiKey(UUID accountId, UUID apiKeyId) {
        final String ctx = CLASSNAME + ".getApiKey";
        try {
            ApiKey apiKey = apiKeyRepository.findByIdAndAccountId(apiKeyId, accountId)
                .orElseThrow(() -> new ApiKeyNotFoundException("API key not found"));
            return apiKeyMapper.toDto(apiKey);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    public Page<ApiKeyResponseDTO> listApiKeys(UUID accountId, Pageable pageable) {
        final String ctx = CLASSNAME + ".listApiKeys";
        try {
            return apiKeyRepository.findByAccountId(accountId, pageable).map(apiKeyMapper::toDto);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }


    public void deleteApiKey(UUID accountId, UUID apiKeyId) {
        final String ctx = CLASSNAME + ".deleteApiKey";
        try {
            ApiKey apiKey = apiKeyRepository.findByIdAndAccountId(apiKeyId, accountId)
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
    public void logUsage(ApiKeyUsageLogIndexDocument apiKeyUsageLog) {
        final String ctx = CLASSNAME + ".logUsage";
        try {
            apiUsageLogRepository.save(apiKeyUsageLog);
        } catch (Exception e) {
            log.error(ctx + ": {}", e.getMessage());
        }
    }

    public Optional<ApiKey> findOneByApiKey(String apiKey) {
        return apiKeyRepository.findOneByApiKey(apiKey);
    }

    public SearchHitsResponseDTO<ApiKeyUsageLogIndexDocument> getApiKeyUsageLogs(User user,
                                                                                 ApiKeyUsageLogCriteria criteria,
                                                                                 Pageable pageable) {
        final String ctx = CLASSNAME + ".getApiKeyUsageLogs";
        try {
            CriteriaQuery query = new CriteriaQuery(criteria != null ? criteria.toCriteriaQuery() : new Criteria(), pageable);
            query.addCriteria(new Criteria(ApiKeyUsageLog_.accountId).is(user.getAccountId()));
            SearchHits<ApiKeyUsageLogIndexDocument> result = elasticsearchOperations.search(query, ApiKeyUsageLogIndexDocument.class);
            return SearchHitsResponseDTO.<ApiKeyUsageLogIndexDocument>builder()
                .totalHits(result.getTotalHits())
                .items(result.stream()
                    .map(SearchHit::getContent)
                    .collect(Collectors.toList()))
                .build();
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    @Scheduled(cron = "0 0 9 * * ?")
    public void checkExpiringApiKeys() {
        Instant fiveDaysFromNow = Instant.now().plus(5, ChronoUnit.DAYS);
        Instant now = Instant.now();
        List<ApiKey> expiringKeys = apiKeyRepository.findAllByExpiresAtAfterAndExpiresAtLessThanEqual(now, fiveDaysFromNow);

        if (!expiringKeys.isEmpty()) {
            Map<UUID, List<ApiKey>> expiringKeysByAccount = expiringKeys.stream()
                .collect(Collectors.groupingBy(ApiKey::getAccountId));

            expiringKeysByAccount.forEach((accountId, apiKeys) -> {
                var principal = userRepository.findByAccountIdAndAccountOwnerIsTrue(accountId.toString()).orElse(null);
                if (principal == null) {
                    return;
                }
                mailService.sendKeyExpirationEmail(principal, apiKeys);

                userNotificationService.createAndSendNotification(principal.getUuid(),
                    NotificationMessageKeyEnum.API_KEY_EXPIRATION,
                    Map.of("names", apiKeys.stream().map(ApiKey::getName).collect(Collectors.joining(","))));
            });
        }
    }
}
