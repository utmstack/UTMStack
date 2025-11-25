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
            if (criteria.getEntityId() != null && criteria.getEntityId().getContains() != null) {
                predicate = cb.and(predicate, cb.like(root.get("entityId"), "%" + criteria.getEntityId().getContains() + "%"));
            }
            if (criteria.getSsoUrl() != null && criteria.getSsoUrl().getContains() != null) {
                predicate = cb.and(predicate, cb.like(root.get("ssoUrl"), "%" + criteria.getSsoUrl().getContains() + "%"));
            }
            if (criteria.getSloUrl() != null && criteria.getSloUrl().getContains() != null) {
                predicate = cb.and(predicate, cb.like(root.get("sloUrl"), "%" + criteria.getSloUrl().getContains() + "%"));
            }
            if (criteria.getCertPem() != null && criteria.getCertPem().getContains() != null) {
                predicate = cb.and(predicate, cb.like(root.get("certPem"), "%" + criteria.getCertPem().getContains() + "%"));
            }
            if (criteria.getNameIdFormat() != null && criteria.getNameIdFormat().getContains() != null) {
                predicate = cb.and(predicate, cb.like(root.get("nameIdFormat"), "%" + criteria.getNameIdFormat().getContains() + "%"));
            }
            if (criteria.getBinding() != null && criteria.getBinding().getContains() != null) {
                predicate = cb.and(predicate, cb.like(root.get("binding"), "%" + criteria.getBinding().getContains() + "%"));
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
