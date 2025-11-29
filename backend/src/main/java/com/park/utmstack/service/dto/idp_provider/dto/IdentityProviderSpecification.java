package com.park.utmstack.service.dto.idp_provider.dto;

import com.park.utmstack.domain.idp_provider.IdentityProviderConfig;
import org.springframework.data.jpa.domain.Specification;

import javax.persistence.criteria.Predicate;

/**
 * Specification builder for IdentityProviderConfig.
 * Adapted for SAML providers only.
 */
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
            if (criteria.getCreatedDate() != null && criteria.getCreatedDate().getEquals() != null) {
                predicate = cb.and(predicate, cb.equal(root.get("createdDate"), criteria.getCreatedDate().getEquals()));
            }
            if (criteria.getLastModifiedDate() != null && criteria.getLastModifiedDate().getEquals() != null) {
                predicate = cb.and(predicate, cb.equal(root.get("lastModifiedDate"), criteria.getLastModifiedDate().getEquals()));
            }
            if (criteria.getActive() != null && criteria.getActive().getEquals() != null) {
                predicate = cb.and(predicate, cb.equal(root.get("active"), criteria.getActive().getEquals()));
            }

            return predicate;
        };
    }
}
