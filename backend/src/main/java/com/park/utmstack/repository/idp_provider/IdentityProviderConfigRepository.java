package com.park.utmstack.repository.idp_provider;

import com.park.utmstack.domain.idp_provider.IdentityProviderConfig;
import com.park.utmstack.domain.idp_provider.enums.ProviderType;
import org.springframework.data.jpa.repository.JpaRepository;

import java.util.List;
import java.util.Optional;

public interface IdentityProviderConfigRepository extends JpaRepository<IdentityProviderConfig, Long> {

    Optional<IdentityProviderConfig> findByProviderTypeAndActiveTrue(ProviderType providerType);

    List<IdentityProviderConfig> findAllByActiveTrue();
}
