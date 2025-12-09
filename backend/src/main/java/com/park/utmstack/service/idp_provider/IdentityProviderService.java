package com.park.utmstack.service.idp_provider;


import com.park.utmstack.domain.idp_provider.IdentityProviderConfig;
import com.park.utmstack.repository.idp_provider.IdentityProviderConfigRepository;
import com.park.utmstack.service.dto.idp_provider.dto.*;
import com.park.utmstack.util.CipherUtil;
import com.park.utmstack.util.events.ProviderChangedEvent;
import com.park.utmstack.util.exceptions.IdpNotFoundException;
import com.park.utmstack.util.exceptions.SamlMetadataUrlInvalidException;
import lombok.RequiredArgsConstructor;
import org.springframework.context.ApplicationEventPublisher;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.domain.Specification;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.io.IOException;
import java.net.HttpURLConnection;
import java.net.URL;
import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;
import java.util.stream.Collectors;

@Service
@RequiredArgsConstructor
public class IdentityProviderService {

    private final IdentityProviderMapper mapper;
    private final IdentityProviderConfigRepository repository;
    private final ApplicationEventPublisher publisher;

    public List<IdentityProviderConfig> getAllActiveProviders() {
        return repository.findAllByActiveTrue();
    }

    public IdentityProviderConfigResponseDto create(IdentityProviderCreateConfigDto dto) {

        validateMetadataUrl(dto.getMetadataUrl());
        IdentityProviderConfig entity = mapper.toEntity(dto);
        entity.setCreatedAt(LocalDateTime.now());
        entity.setUpdatedAt(LocalDateTime.now());
        entity.setSpEntityId(dto.getSpEntityId());
        entity.setSpAcsUrl(dto.getSpAcsUrl());
        IdentityProviderConfig saved = repository.save(entity);
        publisher.publishEvent(new ProviderChangedEvent(saved));
        return mapper.toDto(saved);
    }


    public IdentityProviderConfigResponseDto update(Long id, IdentityProviderConfigRequestDto dto) {

        validateMetadataUrl(dto.getMetadataUrl());

        IdentityProviderConfig existing = repository.findById(id)
                .orElseThrow(() -> new IdpNotFoundException("IdentityProviderConfig not found: " + id));


        existing.setName(dto.getName());
        existing.setMetadataUrl(dto.getMetadataUrl());
        existing.setActive(dto.getActive());
        existing.setUpdatedAt(LocalDateTime.now());

        if(dto instanceof IdentityProviderCreateConfigDto createDto){
            if (createDto.getSpPrivateKeyPem() != null) {
                String encryptedKey = CipherUtil.encrypt(createDto.getSpPrivateKeyPem(), System.getenv("ENCRYPTION_KEY"));
                existing.setSpPrivateKeyPem(encryptedKey);
            }
            if (createDto.getSpCertificatePem() != null) {
                existing.setSpCertificatePem(createDto.getSpCertificatePem());
            }
        }


        IdentityProviderConfig updated = repository.save(existing);
        publisher.publishEvent(new ProviderChangedEvent(updated));
        return mapper.toDto(updated);
    }


    @Transactional(readOnly = true)
    public Page<IdentityProviderConfigResponseDto> findAll(IdentityProviderCriteria criteria, Pageable pageable) {
        Specification<IdentityProviderConfig> spec = IdentityProviderSpecification.build(criteria);
        Page<IdentityProviderConfig> result = repository.findAll(spec, pageable);
        return result.map(mapper::toDto);
    }



    @Transactional(readOnly = true)
    public Optional<IdentityProviderConfigResponseDto> findById(Long id) {
        return repository.findById(id)
                .map(mapper::toDto);
    }


    public void delete(Long id) {
        if (!repository.existsById(id)) {
            throw new IdpNotFoundException("IdentityProviderConfig not found: " + id);
        }
        repository.deleteById(id);
    }

    private void validateMetadataUrl(String metadataUrl) {
        if (metadataUrl == null || metadataUrl.trim().isEmpty()) {
            throw new SamlMetadataUrlInvalidException("Metadata URL is required");
        }

        try {
            URL url = new URL(metadataUrl);
            HttpURLConnection connection = (HttpURLConnection) url.openConnection();
            connection.setRequestMethod("GET");
            connection.setConnectTimeout(5000);
            connection.setReadTimeout(5000);

            int responseCode = connection.getResponseCode();
            if (responseCode != 200) {
                throw new SamlMetadataUrlInvalidException("Metadata URL is not accessible");
            }
        } catch (IOException e) {
            throw new SamlMetadataUrlInvalidException("Failed to access metadata URL");
        }
    }

}
