# O365_GRAPH guide screenshots

Drop one PNG per step (see `M365Guide.tsx` -> `GRAPH_STEPS`):

| File | Step |
|---|---|
| graph-portal.png | Open Microsoft Entra ID |
| graph-app-registrations.png | Go to App registrations |
| graph-register-app.png | Register a new app (UTMStack Graph Dev, single-tenant) |
| graph-certificate-secret.png | Create a client secret |
| graph-api-permission.png | Add a permission -> Microsoft Graph -> Application |
| graph-permissions-list.png | The 4 perms: SecurityEvents.Read.All, Directory.Read.All, AuditLog.Read.All, IdentityRiskEvent.Read.All |
| graph-admin-consent.png | Grant admin consent |
| graph-overview-client.png | Overview: Application (client) ID + Directory (tenant) ID |

Until the PNGs exist the guide renders fine: `StepImg` hides missing images.
