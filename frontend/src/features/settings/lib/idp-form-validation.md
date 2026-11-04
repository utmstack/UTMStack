# Reglas de validación — Formulario de Identity Provider

Documento de referencia de las validaciones del frontend que se aplican sobre el
formulario **Add / Edit identity provider** (`/settings/identity-providers`).

- **Dónde corre:** `src/features/settings/lib/idp-form-validation.ts`
  (`validateIdpForm`), consumido por `UpsertDialog` en
  `src/features/settings/pages/IdentityProvidersPage.tsx`.
- **Cuándo:** en vivo, en cada render del formulario (los errores se pintan bajo
  cada campo y bloquean el botón de guardar). No es asincrónico: no hace peticiones.
- **Qué devuelve:** un objeto `campo → clave i18n` (namespace `idp.form.errors.*`).
  Objeto vacío = formulario válido. El componente traduce la clave al texto.

## Por qué existe

El backend solo comprueba **presencia** al guardar y **formato** al hacer login
(`modules/iam/usecase/idp.go` → `prepareSettings` y `buildSP`/`oidcConfig`/`ldapBind`).
Sin validación en el frontend, un admin puede guardar una URL mal escrita, un PEM
roto o un puerto fuera de rango y el error le llega a un usuario cuando intenta
iniciar sesión, no al admin cuando configura. Este módulo intercede ese gap.

> **La validación de formato aquí es de primera línea, no única.** El backend
> sigue siendo la última defensa y hace el parseo real (X.509 / PKCS8, discovery
> OIDC, dial LDAP) en el momento del login.

## Semántica del secreto (aplica a los tres protocolos)

Cada protocolo lleva exactamente **un campo secreto write-only** que **no** vuelve
del backend:

| Protocolo | Campo secreto |
|---|---|
| saml | `spPrivateKeyPem` |
| oidc | `clientSecret` |
| ldap | `bindPassword` |

Regla:

- **Create** → el secreto es **obligatorio** (sin él no hay nada que guardar).
- **Edit** → vacío significa **"mantener el guardado"**; no se valida formato.
- Si tiene valor en create, **sí** se valida su formato (armadura PEM, en SAML).

## Reglas comunes

| Campo | Tipo | Qué maneja | Validación | Error key |
|---|---|---|---|---|
| `name` | `string` | Nombre del provider; va en la URL de SSO (`/api/v1/sso/<name>/login`) | Obligatorio. Solo en **create**: ≤ 64 chars y `^[A-Za-z0-9][A-Za-z0-9._-]*$` (sin espacios). En **edit** el input está deshabilitado, así que su formato no bloquea guardar | `idp.form.errors.required` / `idp.form.errors.nameFormat` |

En edit, `providerType` también es inmutable (no validable por el usuario).

## SAML

| Campo | Tipo | Qué maneja | Validación | Error key |
|---|---|---|---|---|
| `metadataUrl` | `string` | URL de metadatos del IdP | Obligatorio. URL `http(s)` válida (parseo con `URL`) | `required` / `idp.form.errors.url` |
| `spEntityId` | `string` | Entity ID del Service Provider | Obligatorio. Solo en **create**: URI válida (http, https o `urn:`) — coge espacios y esquemas raros que el IdP rechazaría en login. En **edit** solo presencia (un valor heredado no bloquea el save) | `required` / `idp.form.errors.entityId` |
| `spAcsUrl` | `string` | Endpoint ACS que recibe la respuesta SAML | Obligatorio. URL `http(s)` válida | `required` / `idp.form.errors.url` |
| `spCertificatePem` | `string` | Certificado del SP | Obligatorio. Contiene armadura `-----BEGIN/END CERTIFICATE-----` | `required` / `idp.form.errors.pemCertificate` |
| `spPrivateKeyPem` | `string` | Clave privada del SP (secreto) | Ver sección de secreto | `required` / `idp.form.errors.pemKey` |

## OIDC

| Campo | Tipo | Qué maneja | Validación | Error key |
|---|---|---|---|---|
| `issuer` | `string` | Issuer contra el que corre la discovery | Obligatorio. URL **`https`** válida (https estricto: la discovery expone el secreto) | `required` / `idp.form.errors.httpsUrl` |
| `clientId` | `string` | Client ID de la app | Obligatorio (solo presencia) | `idp.form.errors.required` |
| `redirectUrl` | `string` | Callback registrado en el provider | Obligatorio. URL **https** válida; excepción: loopback `http` (`localhost`, `127.0.0.1`, `[::1]`) para clientes nativos, según RFC 8252 §7.3 — el token exchange no sale de la máquina | `required` / `idp.form.errors.redirectUrl` |
| `clientSecret` | `string` | Client secret (secreto) | Ver sección de secreto | `idp.form.errors.required` |

## LDAP

| Campo | Tipo | Qué maneja | Validación | Error key |
|---|---|---|---|---|
| `host` | `string` | Host del servidor LDAP | Obligatorio. Hostname o IP **sin esquema** ni ruta: `^(?:(?!.*[\/\s])[A-Za-z0-9.-]+|\[[0-9A-Fa-f:]+\])$`. Acepta IPv6 bracketed (`[::1]`) — ver "Paridad con el backend" | `required` / `idp.form.errors.hostname` |
| `port` | `number` | Puerto de conexión | Opcional en formato. Si está, entero sin ceros a la izquierda y ≤ 65535. `0` es válido = default 389 (el backend diala `0` como `389`) | `idp.form.errors.port` |
| `bindDn` | `string` | Servicio que busca en el directorio | Obligatorio (solo presencia; no se valida la sintaxis DN, sería brittle) | `idp.form.errors.required` |
| `baseDn` | `string` | Raíz de la búsqueda | Obligatorio (solo presencia) | `idp.form.errors.required` |
| `userFilter` | `string` | Filtro LDAP donde entra la dirección de login | Obligatorio. Debe contener `%s` (placeholder). Paréntesis balanceados | `required` / `idp.form.errors.userFilterPlaceholder` / `idp.form.errors.userFilterUnbalanced` |
| `bindPassword` | `string` | Password del servicio (secreto) | Ver sección de secreto | `idp.form.errors.required` |

Los atributos LDAP `emailAttribute`, `nameAttribute`, `groupAttribute` y el
toggle `startTls` **no** se validan: el backend los tolera vacíos / usa defaults.

## Decisiones de diseño

- **PEM por armadura, no por parseo.** Se verifica la presencia de las líneas
  `-----BEGIN/END …-----`. Un admin que pega una clave/cert real siempre trae los
  marcadores, y su ausencia es justo el error de dedo que esto debe atrapar. No se
  usa `crypto.subtle` para parsear porque no hay API de "parsear sin usar la clave"
  y solo engordaría la lib. El parseo real (X.509 / PKCS8) lo hace el backend al
  login.
- **`issuer` exige `https`**, mientras que `metadataUrl`/`redirectUrl` aceptan
  `http`: la discovery OIDC viaja al issuer con las credenciales, así que el
  canal debe ser cifrado.
- **`name` estricto solo en create.** En edit el input está deshabilitado; validar
  su formato ahí bloquearía guardar cualquier otro cambio sobre un nombre heredado.
- **`port = 0` es válido.** Borrar el campo produce `0`, que es el default, no un
  typo; no debe atrancar el form.

## Paridad con el backend

Las reglas de formato de esta lib son una primera línea; el backend las repite
en el save (no solo en login) y es la última defensa:

- `backend/modules/iam/usecase/idp.go` → `prepareSettings` (líneas 94, 118, 145)
  llama a `validateSAMLFormat` / `validateOIDCFormat` / `validateLDAPFormat`
  (definidos en `idp_validation.go`) **antes** de cifrar y guardar.
- Reglas idénticas en ambos lados: `name` (mismo `^[A-Za-z0-9][A-Za-z0-9._-]*$`),
  `metadataUrl`/`spAcsUrl` (http URL), `spEntityId` (URI, incluye `urn:`),
  PEM armadura cert, `issuer` (https estricto), `redirectUrl` (https + loopback
  `localhost`/`127.0.0.1`/`::1`), `host` (sin esquema), `port` (≤ 65535,
  `0` aceptado), `userFilter` (placeholder + paréntesis balanceados).

**Divergencia conocida — IPv6 brackets (frontend más permisivo):**

- El frontend acepta `host = [::1]` (`HOST_RE` con la alternativa
  `\[[0-9A-Fa-f:]+\]`).
- El backend **no** lo acepta aún: `idpHostRe` es la misma regex sin brackets
  (`idp_validation.go:16`) y el dial construye la URL con
  `fmt.Sprintf("ldap://%s:%d", host, port)` (`idp.go:880`), que con un IPv6
  necesitaría `net.JoinHostPort` para bracketear.
- Consecuencia: un admin puede llenar el form con `[::1]` sin error local, pero
  el save responde `ErrIDPSettingsInvalid` (mensaje genérico del backend).
  Es un trade-off aceptado: la UX del form no bloquea hosts IPv6 y la
  compatibilidad end-to-end queda pendiente de un cambio de backend (regex +
  dial). Se documenta aquí y en el comentario sobre `HOST_RE` en el `.ts`.

**Nota del backend (no se toca desde esta rama):** `validateIDPSettingsFormat`
(`idp_validation.go:19`) no tiene caller — las funciones individuales son las
que corren en `prepareSettings`.

## Fuentes de las reglas

Las reglas de formato (URL, `https` estricto para `issuer`, loopback para
`redirectUrl`, URI para `spEntityId`, placeholder `%s` y paréntesis para
`userFilter`) se cruzaron contra las specs de cada protocolo:

- **OIDC** — [OpenID Connect Core 1.0 §3.1.2.1](https://openid.net/specs/openid-connect-core-1_0.html)
  (`client_id`, `redirect_uri` required en la auth request) y
  [RFC 8252 §7.3](https://datatracker.ietf.org/doc/html/rfc8252#section-7.3)
  (loopback redirect URIs sobre http para clientes nativos).
- **SAML** — guías de proveedores (Cisco IPR, OneStream, Vendasta): el entity ID
  del SP puede ser `http(s)` o `urn:`; ACS y metadata siempre son URLs
  navegables, y la cert/keys son X.509 / PEM.
- **LDAP** — flujo `service bind → search → user bind` del propio backend
  (`modules/iam/usecase/idp.go`): por eso `bindDn` y `bindPassword` son
  obligatorios aquí aunque algunas fuentes los marquen opcionales (simple bind).

El set de *required* por protocolo (presencia) no lo define esta lib, sino el
backend en `prepareSettings`. Esta lib añade el *formato* por encima, de forma
aditiva.

## Cubertura

`src/features/settings/lib/idp-form-validation.test.ts` — 45 casos, uno por
regla, corriendo con `npx vitest run`.
