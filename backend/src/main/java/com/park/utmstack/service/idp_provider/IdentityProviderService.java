package com.park.utmstack.service.idp_provider;


import com.park.utmstack.config.Constants;
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
import java.net.InetAddress;
import java.net.MalformedURLException;
import java.net.UnknownHostException;
import java.net.URL;
import java.time.LocalDateTime;
import java.util.Optional;
import java.util.Set;

@Service
@RequiredArgsConstructor
public class IdentityProviderService {

    private final IdentityProviderMapper mapper;
    private final IdentityProviderConfigRepository repository;
    private final ApplicationEventPublisher publisher;

    public IdentityProviderConfigResponseDto create(IdentityProviderCreateConfigDto dto) {

        validateMetadataUrl(dto.getMetadataUrl());

        // Validate encryption key before mapper encrypts the private key
        getValidatedEncryptionKey();

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
                String encryptionKey = getValidatedEncryptionKey();
                String encryptedKey = CipherUtil.encrypt(createDto.getSpPrivateKeyPem(), encryptionKey);
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

    /**
     * Validates and retrieves the encryption key from environment variables.
     *
     * @return The validated encryption key
     * @throws IllegalStateException if ENCRYPTION_KEY is not configured
     */
    private String getValidatedEncryptionKey() {
        String encryptionKey = System.getenv(Constants.ENV_ENCRYPTION_KEY);
        if (encryptionKey == null || encryptionKey.isBlank()) {
            throw new IllegalStateException(
                    "Environment variable " + Constants.ENV_ENCRYPTION_KEY + " not configured");
        }
        return encryptionKey;
    }

    /**
     * Schemes accepted for SAML metadata endpoints. HTTP is only permitted for the
     * documented in-lab default, see {@link #isMetadataUrlAllowed(URL)}.
     */
    private static final Set<String> ALLOWED_METADATA_SCHEMES = Set.of("http", "https");

    private void validateMetadataUrl(String metadataUrl) {
        if (metadataUrl == null || metadataUrl.trim().isEmpty()) {
            throw new SamlMetadataUrlInvalidException("Metadata URL is required");
        }

        HttpURLConnection connection = null;
        try {
            URL url = new URL(metadataUrl);

            if (!isMetadataUrlAllowed(url)) {
                throw new SamlMetadataUrlInvalidException(
                        "Metadata URL must use the http or https scheme and target a public hostname");
            }

            connection = (HttpURLConnection) url.openConnection();
            connection.setRequestMethod("GET");
            connection.setConnectTimeout(5000);
            connection.setReadTimeout(5000);
            connection.setInstanceFollowRedirects(false);

            int responseCode = connection.getResponseCode();
            if (responseCode != 200) {
                throw new SamlMetadataUrlInvalidException(
                        String.format("Metadata URL is not accessible. HTTP Status: %d", responseCode));
            }
        } catch (MalformedURLException e) {
            throw new SamlMetadataUrlInvalidException(
                    "Invalid metadata URL format: " + e.getMessage());
        } catch (IOException e) {
            throw new SamlMetadataUrlInvalidException(
                    "Failed to access metadata URL: " + e.getMessage());
        } catch (Exception e) {
            throw new SamlMetadataUrlInvalidException(
                    "Unexpected error validating metadata URL: " + e.getMessage());
        } finally {
            if (connection != null) {
                connection.disconnect();
            }
        }
    }

    /**
     * Rejects metadata URLs that cannot be a legitimate SAML IdP endpoint:
     * non-http(s) schemes, literal IP addresses (loopback, private, link-local such
     * as 169.254.169.254), and hostnames that resolve to internal ranges.
     */
    private boolean isMetadataUrlAllowed(URL url) {
        String scheme = url.getProtocol() == null ? "" : url.getProtocol().toLowerCase();
        if (!ALLOWED_METADATA_SCHEMES.contains(scheme)) {
            return false;
        }

        String host = url.getHost();
        if (host == null || host.isEmpty()) {
            return false;
        }

        InetAddress address;
        try {
            // Literal IP in the URL, or DNS resolution of the hostname.
            address = InetAddress.getByName(host);
        } catch (UnknownHostException e) {
            return false;
        }

        return !address.isLoopbackAddress()
                && !address.isSiteLocalAddress()
                && !address.isLinkLocalAddress()
                && !address.isAnyLocalAddress()
                && !address.isMulticastAddress();
    }

}
