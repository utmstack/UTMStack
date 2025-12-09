package com.park.utmstack.service.dto.idp_provider.dto;

import com.park.utmstack.domain.idp_provider.enums.ProviderType;
import com.park.utmstack.validation.saml.ValidCertificate;
import com.park.utmstack.validation.saml.ValidPrivateKey;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.EqualsAndHashCode;
import lombok.NoArgsConstructor;

import javax.validation.constraints.NotBlank;
import javax.validation.constraints.NotNull;

@EqualsAndHashCode(callSuper = true)
@Data
@NoArgsConstructor
@AllArgsConstructor
public class IdentityProviderCreateConfigDto extends IdentityProviderConfigRequestDto {

    @NotBlank
    @ValidPrivateKey
    private String spPrivateKeyPem;

    @NotBlank
    private String spCertificatePem;

}