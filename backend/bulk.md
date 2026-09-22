# Bulk Cross-Tenant Endpoints — Summary

All routes require `userAuth + RequirePlatform() + RequirePermission("<module>.write")`. Body always carries `selector: {tenantIds: [], allTenants: bool}` plus the resource payload. Response: `{succeeded: [tenantId...], failed: [{tenantId, error}...]}`.

## eventprocessing (pipelines)

| Method | Path | What it does |
|---|---|---|
| POST | `/platform/eventprocessing/pipelines/bulk/create` | Install the same pipeline (relPath + YAML content) as a user overlay in every selected tenant |
| POST | `/platform/eventprocessing/pipelines/bulk/update` | Rewrite a user pipeline's content across tenants (system pipelines refuse) |
| POST | `/platform/eventprocessing/pipelines/bulk/delete` | Remove a user pipeline across tenants (system pipelines refuse) |
| POST | `/platform/eventprocessing/pipelines/bulk/activate` | Enable/disable a pipeline per tenant (allowed on system pipelines — toggle only) |

## eventprocessing (correlation rules)

| Method | Path | What it does |
|---|---|---|
| POST | `/platform/eventprocessing/correlation-rule/bulk/create` | Install the same correlation rule (full DTO) as a user overlay in every selected tenant |
| POST | `/platform/eventprocessing/correlation-rule/bulk/update` | Update a correlation rule by relPath across tenants (system-owned refuse) |
| POST | `/platform/eventprocessing/correlation-rule/bulk/delete` | Delete a correlation rule by relPath across tenants (system-owned refuse) |
| POST | `/platform/eventprocessing/correlation-rule/bulk/activate` | Enable/disable a correlation rule per tenant |

## soar rules (flujos / playbooks)

| Method | Path | What it does |
|---|---|---|
| POST | `/platform/soar/rules/bulk/create` | Deploy the same SOAR flow YAML to every selected tenant |
| POST | `/platform/soar/rules/bulk/update` | Update a flow's content across tenants (product-shipped flows refuse) |
| POST | `/platform/soar/rules/bulk/delete` | Delete a flow across tenants (product-shipped flows refuse) |
| POST | `/platform/soar/rules/bulk/enable` | Toggle a flow enabled/disabled per tenant (works on system flows) |

## compliance

| Method | Path | What it does |
|---|---|---|
| POST | `/platform/compliance/frameworks/bulk/create` | Add the same custom framework (PCI/HIPAA-style) as a user overlay in every selected tenant |
| POST | `/platform/compliance/frameworks/bulk/update` | Update a framework across tenants (vendor/system frameworks refuse) |
| POST | `/platform/compliance/frameworks/bulk/delete` | Remove a framework overlay across tenants (system-shipped refuse) |
| POST | `/platform/compliance/controls/bulk/create` | Add the same custom control across tenants |
| POST | `/platform/compliance/controls/bulk/update` | Update a control across tenants (system controls refuse) |
| POST | `/platform/compliance/controls/bulk/delete` | Delete a control overlay across tenants (system controls refuse) |

## SMTP (appconfig)

| Method | Path | What it does |
|---|---|---|
| POST | `/platform/config/smtp/bulk/update` | Upsert one or more `utmstack.mail.*` fields (host, port, user, password, from, authType, orgname, baseUrl) across tenants; password is encrypted before storage |
| POST | `/platform/config/smtp/bulk/test` | Fire a self-ping test email per tenant with the given SMTP config; reports which tenants delivered vs. failed |

*`allTenants=true` deliberately excludes `DefaultTenantID` so the platform-plane SMTP isn't overwritten.*

## IDP (iam) — also gated by `enterprise`

| Method | Path | What it does |
|---|---|---|
| POST | `/platform/identity-providers/bulk/create` | Register the same IDP (SAML/OIDC/LDAP) in every selected tenant; secrets (clientSecret, privateKey, bindPassword) encrypted at rest |
| POST | `/platform/identity-providers/bulk/update` | Update an IDP config across tenants (omitted secret fields keep prior value) |
| POST | `/platform/identity-providers/bulk/delete` | Remove an IDP across tenants |

## branding (appconfig) — also gated by `enterprise`

| Method | Path | What it does |
|---|---|---|
| POST | `/platform/branding/bulk/update` | Apply the same branding JSON (product name, accent color, language, logo URLs, enabled flag) across tenants |
| POST | `/platform/branding/bulk/upload-asset/:slot` | Upload one asset file once (`logo`, `logoDark`, `favicon`, `reportLogo`, `reportCover`), point every selected tenant's branding to the resulting URL |

*`allTenants=true` also excludes `DefaultTenantID` for branding.*

---

**Response shape (all endpoints)**
```json
{
  "succeeded": ["tenantId-a", "tenantId-b"],
  "failed":    [{"tenantId": "tenantId-c", "error": "system-owned pipeline cannot be updated"}]
}
```

Partial success is normal — a system-owned refusal, an enterprise-license failure, or an invalid payload on one tenant does NOT stop the loop, it lands in `failed` while the rest of the tenants continue.
