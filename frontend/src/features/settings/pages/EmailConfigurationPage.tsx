import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Check, Eye, EyeOff, Loader2, Mail, Send, X } from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { EmailChipInput } from '@/shared/components/ui/email-chip-input'
import { PlatformBroadcastButton, broadcast, BULK_PATHS } from '@/features/platform-broadcast'
import { configHttpService } from '../services/config-http.service'

/* Backend config keys (utm_configuration_parameter), read/written via /config/:key. */
const K = {
  host: 'utmstack.mail.host',
  port: 'utmstack.mail.port',
  username: 'utmstack.mail.username',
  password: 'utmstack.mail.password',
  from: 'utmstack.mail.from',
  baseUrl: 'utmstack.mail.baseUrl',
  organization: 'utmstack.mail.organization',
  auth: 'utmstack.mail.properties.mail.smtp.auth',
  // Alerts and incidents keep separate lists: an incident is raised by a person
  // and there are few, while alerts arrive on their own at whatever rate the
  // environment produces. One shared list means whoever wants incident mail
  // also gets every alert, and ends up filtering both away.
  alertsTo: 'utmstack.alerts.notification_to',
  alertsCc: 'utmstack.alerts.notification_cc',
  incidentsTo: 'utmstack.incidents.notification_to',
  incidentsCc: 'utmstack.incidents.notification_cc',
} as const

const splitCsv = (s: string): string[] =>
  s.split(',').map((x) => x.trim()).filter(Boolean)

interface Recipients {
  to: string[]
  cc: string[]
}

const NO_RECIPIENTS: Recipients = { to: [], cc: [] }

const sameList = (a: string[], b: string[]) => a.length === b.length && a.every((x, i) => x === b[i])
const sameRecipients = (a: Recipients, b: Recipients) => sameList(a.to, b.to) && sameList(a.cc, b.cc)

type Encryption = 'TLS' | 'SSL' | 'NONE'
const ENCRYPTIONS: Encryption[] = ['TLS', 'SSL', 'NONE']

interface Form {
  host: string
  port: string
  username: string
  password: string
  from: string
  baseUrl: string
  organization: string
  encryption: Encryption
}

const EMPTY: Form = {
  host: '',
  port: '587',
  username: '',
  password: '',
  from: '',
  baseUrl: '',
  organization: '',
  encryption: 'TLS',
}

export function EmailConfigurationPage() {
  const { t } = useTranslation()
  const [form, setForm] = useState<Form>(EMPTY)
  const [initial, setInitial] = useState<Form>(EMPTY)
  const [passwordSet, setPasswordSet] = useState(false)
  const [passwordTouched, setPasswordTouched] = useState(false)
  const [showPass, setShowPass] = useState(false)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [test, setTest] = useState<'idle' | 'sending' | 'success' | 'fail'>('idle')

  // Recipient lists — each persisted as a comma-separated string under its key.
  const [alerts, setAlerts] = useState<Recipients>(NO_RECIPIENTS)
  const [initialAlerts, setInitialAlerts] = useState<Recipients>(NO_RECIPIENTS)
  const [incidents, setIncidents] = useState<Recipients>(NO_RECIPIENTS)
  const [initialIncidents, setInitialIncidents] = useState<Recipients>(NO_RECIPIENTS)


  useEffect(() => {
    let cancelled = false
    configHttpService
      .list()
      .then((entries) => {
        if (cancelled) return
        const m = new Map(entries.map((e) => [e.key, e]))
        const val = (k: string) => m.get(k)?.value ?? ''
        const v: Form = {
          host: val(K.host),
          port: val(K.port) || '587',
          username: val(K.username),
          password: '',
          from: val(K.from),
          baseUrl: val(K.baseUrl),
          organization: val(K.organization),
          encryption: (val(K.auth) as Encryption) || 'TLS',
        }
        setForm(v)
        setInitial(v)
        setPasswordSet(!!m.get(K.password)?.is_set)
        const a: Recipients = { to: splitCsv(val(K.alertsTo)), cc: splitCsv(val(K.alertsCc)) }
        const i: Recipients = { to: splitCsv(val(K.incidentsTo)), cc: splitCsv(val(K.incidentsCc)) }
        setAlerts(a)
        setInitialAlerts(a)
        setIncidents(i)
        setInitialIncidents(i)
      })
      .catch(() => {
        if (!cancelled) toast.error(t('emailConfig.loadError'))
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [t])

  const patch = (p: Partial<Form>) => {
    setForm((f) => ({ ...f, ...p }))
    setTest('idle')
  }

  const dirty =
    passwordTouched ||
    (Object.keys(initial) as (keyof Form)[]).some((k) => k !== 'password' && form[k] !== initial[k]) ||
    !sameRecipients(alerts, initialAlerts) ||
    !sameRecipients(incidents, initialIncidents)

  // One save for the page. Every changed key goes in the same batch, and only
  // the changed ones: rewriting a key nobody touched stamps its updated_at for
  // no reason.
  const save = async () => {
    setSaving(true)
    try {
      const updates: Promise<unknown>[] = []
      const put = (key: string, value: string) => updates.push(configHttpService.set(key, { value }))
      if (form.host !== initial.host) put(K.host, form.host)
      if (form.port !== initial.port) put(K.port, form.port)
      if (form.username !== initial.username) put(K.username, form.username)
      if (form.from !== initial.from) put(K.from, form.from)
      if (form.baseUrl !== initial.baseUrl) put(K.baseUrl, form.baseUrl)
      if (form.organization !== initial.organization) put(K.organization, form.organization)
      if (form.encryption !== initial.encryption) put(K.auth, form.encryption)
      // Only write the password when the admin actually typed a new one (a blank
      // field keeps the stored secret).
      if (passwordTouched) put(K.password, form.password)

      const putList = (key: string, next: string[], was: string[]) => {
        if (!sameList(next, was)) put(key, next.join(','))
      }
      putList(K.alertsTo, alerts.to, initialAlerts.to)
      putList(K.alertsCc, alerts.cc, initialAlerts.cc)
      putList(K.incidentsTo, incidents.to, initialIncidents.to)
      putList(K.incidentsCc, incidents.cc, initialIncidents.cc)

      await Promise.all(updates)
      setInitialAlerts(alerts)
      setInitialIncidents(incidents)
      setInitial({ ...form, password: '' })
      if (passwordTouched && form.password) setPasswordSet(true)
      setPasswordTouched(false)
      setForm((f) => ({ ...f, password: '' }))
      toast.success(t('emailConfig.saved'))
    } catch {
      toast.error(t('emailConfig.saveError'))
    } finally {
      setSaving(false)
    }
  }

  const runTest = async () => {
    if (!form.from) {
      toast.error(t('emailConfig.test.fromRequired'))
      return
    }
    setTest('sending')
    try {
      const res = await configHttpService.checkMail([
        {
          host: form.host,
          port: Number(form.port) || 0,
          username: form.username,
          password: passwordTouched ? form.password : '',
          authType: form.encryption,
          from: form.from,
        },
      ])
      if (res.success) {
        setTest('success')
        toast.success(t('emailConfig.test.success', { email: form.from }))
      } else {
        setTest('fail')
        toast.error(res.message || t('emailConfig.test.fail'))
      }
    } catch (err) {
      setTest('fail')
      toast.error(err instanceof Error ? err.message : t('emailConfig.test.fail'))
    }
  }

  const smtpPayload = {
    host: form.host,
    port: form.port,
    username: form.username,
    password: form.password,
    from: form.from,
    authType: form.encryption,
    orgname: form.organization,
    baseUrl: form.baseUrl,
  }

  const onBroadcastUpdate = async (selector: { tenantIds: string[]; allTenants: boolean }) => {
    return broadcast(BULK_PATHS.smtp.update, selector, smtpPayload)
  }

  return (
    <div className="w-full px-6 pb-6 pt-3">
      <header>
        <h1 className="flex items-center gap-2 text-base font-semibold">
          <Mail size={16} strokeWidth={1.75} />
          {t('emailConfig.title')}
        </h1>
      </header>

      {loading ? (
        <div className="mt-6 flex items-center gap-2 text-sm text-muted-foreground">
          <Loader2 className="h-4 w-4 animate-spin" />
        </div>
      ) : (
        <div className="mt-6 space-y-5">
          {/* SMTP server */}
          <Section title={t('emailConfig.server.title')}>
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <Field label={t('emailConfig.server.host')} hint={t('emailConfig.server.hostHint')}>
                <Input value={form.host} onChange={(e) => patch({ host: e.target.value })} placeholder="smtp.example.com" />
              </Field>
              <Field label={t('emailConfig.server.port')}>
                <Input
                  type="number"
                  value={form.port}
                  onChange={(e) => patch({ port: e.target.value })}
                  placeholder="587"
                  className="max-w-[140px]"
                />
              </Field>
              <Field label={t('emailConfig.server.encryption')}>
                <div className="inline-flex rounded-md border border-border p-0.5">
                  {ENCRYPTIONS.map((enc) => (
                    <button
                      key={enc}
                      type="button"
                      onClick={() => patch({ encryption: enc })}
                      className={cn(
                        'rounded px-3 py-1 text-xs transition-colors',
                        form.encryption === enc ? 'bg-muted font-medium text-foreground' : 'text-muted-foreground hover:text-foreground',
                      )}
                    >
                      {t(`emailConfig.encryptionOpt.${enc}`)}
                    </button>
                  ))}
                </div>
              </Field>
              <Field label={t('emailConfig.server.username')}>
                <Input value={form.username} onChange={(e) => patch({ username: e.target.value })} autoComplete="off" />
              </Field>
              <Field
                label={t('emailConfig.server.password')}
                hint={passwordSet && !passwordTouched ? t('emailConfig.server.passwordSetHint') : undefined}
              >
                <div className="relative max-w-sm">
                  <Input
                    type={showPass ? 'text' : 'password'}
                    value={form.password}
                    onChange={(e) => {
                      setForm((f) => ({ ...f, password: e.target.value }))
                      setPasswordTouched(true)
                      setTest('idle')
                    }}
                    placeholder={passwordSet ? '••••••••' : ''}
                    autoComplete="new-password"
                    className="pr-9"
                  />
                  <button
                    type="button"
                    onClick={() => setShowPass((v) => !v)}
                    className="absolute right-2 top-1/2 -translate-y-1/2 rounded p-1 text-muted-foreground hover:bg-muted hover:text-foreground"
                  >
                    {showPass ? <EyeOff size={14} /> : <Eye size={14} />}
                  </button>
                </div>
              </Field>
            </div>
          </Section>

          {/* Sender */}
          <Section title={t('emailConfig.sender.title')}>
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <Field label={t('emailConfig.sender.from')} hint={t('emailConfig.sender.fromHint')}>
                <Input type="email" value={form.from} onChange={(e) => patch({ from: e.target.value })} placeholder="alerts@example.com" />
              </Field>
              <Field label={t('emailConfig.sender.organization')} hint={t('emailConfig.sender.organizationHint')}>
                <Input value={form.organization} onChange={(e) => patch({ organization: e.target.value })} placeholder="Acme Corp" />
              </Field>
              <Field label={t('emailConfig.sender.baseUrl')} hint={t('emailConfig.sender.baseUrlHint')}>
                <Input value={form.baseUrl} onChange={(e) => patch({ baseUrl: e.target.value })} placeholder="https://siem.example.com" />
              </Field>
            </div>
          </Section>

          <RecipientsSection
            title={t('emailConfig.recipients.alertsTitle')}
            helper={t('emailConfig.recipients.alertsHelper')}
            value={alerts}
            onChange={setAlerts}
          />

          <RecipientsSection
            title={t('emailConfig.recipients.incidentsTitle')}
            helper={t('emailConfig.recipients.incidentsHelper')}
            value={incidents}
            onChange={setIncidents}
          />

          {/* Actions */}
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div className="flex items-center gap-2">
              <Button variant="outline" size="sm" disabled={test === 'sending'} onClick={() => void runTest()}>
                {test === 'sending' ? (
                  <Loader2 size={14} className="mr-1.5 animate-spin" />
                ) : test === 'success' ? (
                  <Check size={14} className="mr-1.5 text-emerald-500" />
                ) : test === 'fail' ? (
                  <X size={14} className="mr-1.5 text-red-500" />
                ) : (
                  <Send size={14} className="mr-1.5" />
                )}
                {test === 'sending' ? t('emailConfig.test.sending') : t('emailConfig.test.button')}
              </Button>
              {form.from && (
                <span className="text-[11px] text-muted-foreground">{t('emailConfig.test.sentTo', { email: form.from })}</span>
              )}
            </div>
            <div className="flex items-center gap-2">
              <PlatformBroadcastButton
                label={t('platformBroadcast.button')}
                title={t('platformBroadcast.action.update', { resource: t('platformBroadcast.resource.smtp') })}
                excludeDefaultTenant={true}
                onBroadcast={onBroadcastUpdate}
              />
              <Button size="sm" disabled={!dirty || saving} onClick={() => void save()}>
                {saving ? t('emailConfig.saving') : t('emailConfig.save')}
              </Button>
            </div>
          </div>

        </div>
      )}
    </div>
  )
}

/* ─── Small parts ──────────────────────────────────────────────────────── */

/** A To/CC pair for one kind of notification. It has no save button of its own:
 *  the page saves as a whole, so an admin who edits the server and the
 *  recipients presses save once. */
function RecipientsSection({
  title,
  helper,
  value,
  onChange,
}: {
  title: string
  helper: string
  value: Recipients
  onChange: (r: Recipients) => void
}) {
  const { t } = useTranslation()

  return (
    <Section title={title}>
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <Field label={t('emailConfig.recipients.to')}>
          <EmailChipInput
            values={value.to}
            onChange={(to) => onChange({ ...value, to })}
            placeholder="alice@example.com"
            addLabel={t('emailConfig.recipients.add')}
          />
        </Field>
        <Field label={t('emailConfig.recipients.cc')}>
          <EmailChipInput
            values={value.cc}
            onChange={(cc) => onChange({ ...value, cc })}
            placeholder="bob@example.com"
            addLabel={t('emailConfig.recipients.add')}
          />
        </Field>
      </div>
      <p className="mt-3 text-[11px] text-muted-foreground">{helper}</p>
    </Section>
  )
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section className="rounded-xl border border-border bg-card p-6">
      <h2 className="mb-4 text-sm font-semibold">{title}</h2>
      {children}
    </section>
  )
}

function Field({ label, hint, children }: { label: string; hint?: string; children: React.ReactNode }) {
  return (
    <div className="space-y-1.5">
      <label className="block text-xs font-medium text-foreground/80">{label}</label>
      {children}
      {hint && <p className="text-[11px] text-muted-foreground">{hint}</p>}
    </div>
  )
}
