package com.park.utmstack.service.idp_provider;


import com.park.utmstack.domain.idp_provider.IdentityProviderConfig;
import com.park.utmstack.repository.idp_provider.IdentityProviderConfigRepository;
import com.park.utmstack.service.dto.idp_provider.dto.*;
import com.park.utmstack.util.events.ProviderChangedEvent;
import com.park.utmstack.util.exceptions.IdpNotFoundException;
import lombok.RequiredArgsConstructor;
import org.springframework.context.ApplicationEventPublisher;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.domain.Specification;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

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

        IdentityProviderConfig entity = mapper.toEntity(dto);
        entity.setCreatedAt(LocalDateTime.now());
        entity.setUpdatedAt(LocalDateTime.now());

        IdentityProviderConfig saved = repository.save(entity);
        publisher.publishEvent(new ProviderChangedEvent(saved));
        return mapper.toDto(saved);
    }


    public IdentityProviderConfigResponseDto update(Long id, IdentityProviderConfigRequestDto dto) {

        IdentityProviderConfig existing = repository.findById(id)
                .orElseThrow(() -> new IdpNotFoundException("IdentityProviderConfig not found: " + id));

        existing.setName(dto.getName());
        existing.setEntityId(dto.getEntityId());

        existing.setActive(dto.getActive());
        existing.setUpdatedAt(LocalDateTime.now());

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
}
