package com.park.utmstack.service.dto.idp_provider.dto;

import com.park.utmstack.domain.idp_provider.IdentityProviderConfig;
import org.springframework.data.jpa.domain.Specification;

import javax.persistence.criteria.Predicate;

public class IdentityProviderSpecification {
    public static Specification<IdentityProviderConfig> build(IdentityProviderCriteria criteria) {
        return (root, query, cb) -> {
            Predicate predicate = cb.conjunction();

            if (criteria.getId() != null && criteria.getId().getEquals() != null) {
                predicate = cb.and(predicate, cb.equal(root.get("id"), criteria.getId().getEquals()));
            }
            if (criteria.getName() != null && criteria.getName().getContains() != null) {
                predicate = cb.and(predicate, cb.like(root.get("name"), "%" + criteria.getName().getContains() + "%"));
            }
            if (criteria.getProviderType() != null && criteria.getProviderType().getEquals() != null) {
                predicate = cb.and(predicate, cb.equal(root.get("providerType"), criteria.getProviderType().getEquals()));
            }
            if (criteria.getRedirectUri() != null && criteria.getRedirectUri().getContains() != null) {
                predicate = cb.and(predicate, cb.like(root.get("redirectUri"), "%" + criteria.getRedirectUri().getContains() + "%"));
            }
            if (criteria.getScopes() != null && criteria.getScopes().getContains() != null) {
                predicate = cb.and(predicate, cb.like(root.get("scopes"), "%" + criteria.getScopes().getContains() + "%"));
            }
            if (criteria.getAuthUri() != null && criteria.getAuthUri().getContains() != null) {
                predicate = cb.and(predicate, cb.like(root.get("authUri"), "%" + criteria.getAuthUri().getContains() + "%"));
            }
            if (criteria.getTokenUri() != null && criteria.getTokenUri().getContains() != null) {
                predicate = cb.and(predicate, cb.like(root.get("tokenUri"), "%" + criteria.getTokenUri().getContains() + "%"));
            }
            if (criteria.getUserInfoUri() != null && criteria.getUserInfoUri().getContains() != null) {
                predicate = cb.and(predicate, cb.like(root.get("userInfoUri"), "%" + criteria.getUserInfoUri().getContains() + "%"));
            }
            if (criteria.getJwksUri() != null && criteria.getJwksUri().getContains() != null) {
                predicate = cb.and(predicate, cb.like(root.get("jwksUri"), "%" + criteria.getJwksUri().getContains() + "%"));
            }
            if (criteria.getCreatedDate() != null && criteria.getCreatedDate().getEquals() != null) {
                predicate = cb.and(predicate, cb.equal(root.get("createdDate"), criteria.getCreatedDate().getEquals()));
            }
            if (criteria.getLastModifiedDate() != null && criteria.getLastModifiedDate().getEquals() != null) {
                predicate = cb.and(predicate, cb.equal(root.get("lastModifiedDate"), criteria.getLastModifiedDate().getEquals()));
            }

            return predicate;
        };
    }
}

