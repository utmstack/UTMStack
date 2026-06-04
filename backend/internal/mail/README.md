# Mail Provider

SMTP mail provider. Sends plain HTML messages or template-rendered messages, with optional attachments, using configuration persisted in the application config store.

## Layout

```
mail/
├── connectors/   # interfaces (ports)
│   ├── usecase.go        → MailService
│   └── repository.go     → MailConfigurationRepository
├── domain/       # value types
│   ├── email.go          → EmailConfig
│   └── attatchment.go    → Attatchment
├── repository/   # MailConfigurationRepository impl backed by appconfig.Store
└── usecase/      # MailService impl (net/smtp)
```

## Interfaces

### `MailService`

```go
SendMail(ctx, addresses, body, attachments) error
SendTemplateMail(ctx, addresses, template, vars, locale) error
```

- `SendMail` sends a `multipart/mixed` message: the body is `text/html; charset=UTF-8` (quoted-printable), attachments are base64-encoded.
- `SendTemplateMail` parses `template` with `html/template`. Variables are exposed both as top-level keys and under `.Vars`; the locale is available under `.Locale`. The rendered output is forwarded to `SendMail` with no attachments.

### `MailConfigurationRepository`

```go
GetMailConfiguration(ctx) (*EmailConfig, error)
SetMailConfiguration(ctx, EmailConfig) error
```

Reads/writes each field as a separate entry in the application config store (see `shared/constants` for the key names — `PROP_MAIL_HOST`, `PROP_MAIL_PORT`, …). `Password` is stored with `IsSecret: true`.

## Domain

### `EmailConfig`

| Field | Purpose |
| --- | --- |
| `Host`, `Port` | SMTP server address. Required. |
| `Username`, `Password` | Credentials used when `SmtpAuth == "true"`. |
| `SmtpAuth` | String flag (`"true"`/`"false"`). When true, `smtp.PlainAuth` is used. |
| `From` | Sender address. Falls back to `Username` if empty. |
| `Orgname` | Emitted as the `X-Organization` header when present. |
| `BaseUrl` | Base URL for links rendered into templates. |
| `ProtocolValue`, `PortTlsValue`, `PortSslValue`, `PortNoneValue` | Read-only defaults sourced from `shared/constants`, surfaced for UI/config screens. |

### `Attatchment`

`Filename`, `ContentType` (defaults to `application/octet-stream` when empty), and raw `Bytes`.

## Wiring

```go
store := appconfig.NewStore(...)               // existing appconfig.Store
mailRepo := repository.NewMailConfigurationRepository(store)
mail := usecase.New(mailRepo)

err := mail.SendMail(ctx, []string{"user@example.com"}, "<p>hello</p>", nil)
```

## Behavior notes

- `SendMail` fails fast if `Host`/`Port` are empty or `addresses` is empty.
- Auth is skipped entirely when `SmtpAuth != "true"` or `Username` is empty — useful for relays that accept unauthenticated submissions on the internal network.
- Transport is plain `net/smtp.SendMail`; STARTTLS/SSL negotiation is whatever the server advertises during `SendMail`. There is no implicit TLS wrapper.
