import type { TFunction } from 'i18next'
import { Bug, Code2, FileWarning, Globe2, LockKeyhole, Workflow } from 'lucide-react'
import { cn } from '@/shared/lib/utils'

type EdrKind = 'file' | 'network' | 'ransomware' | 'script' | 'process' | 'endpoint'

interface EdrDetailPanelProps {
  flat: Record<string, unknown>
  t: TFunction
}

const text = (flat: Record<string, unknown>, aliases: string[]): string | null => {
  for (const alias of aliases) {
    const value = flat[alias]
    if (value != null && value !== '') return typeof value === 'object' ? JSON.stringify(value) : String(value)
  }
  return null
}

function detectKind(flat: Record<string, unknown>): EdrKind {
  const values = Object.entries(flat).map(([key, value]) => `${key} ${String(value)}`).join(' ').toLowerCase()
  if (/ransom|cryptoguard|encryption|score.?signal|contained/.test(values)) return 'ransomware'
  if (/script|powershell|command.?line|amsi|macro/.test(values)) return 'script'
  if (/network.?block|connection.?block|remote.?address|indicator.?match|dns|domain/.test(values)) return 'network'
  if (/process|parent.?process|pid/.test(values)) return 'process'
  if (/file|malware|quarantine|hash|sha256|md5|verdict/.test(values)) return 'file'
  return 'endpoint'
}

function isEdrEvent(flat: Record<string, unknown>): boolean {
  const values = Object.entries(flat).map(([key, value]) => `${key} ${String(value)}`).join(' ').toLowerCase()
  return /\bedr\b|endpoint|crowdstrike|sentinel.?one|bitdefender|sophos|defender|ransomware|script.?block|network.?block|file.?detect/.test(values)
}

const kindMeta: Record<EdrKind, { icon: typeof Bug; tone: string; title: string }> = {
  file: { icon: FileWarning, tone: 'text-rose-500 bg-rose-500/10 ring-rose-500/20', title: 'file' },
  network: { icon: Globe2, tone: 'text-sky-500 bg-sky-500/10 ring-sky-500/20', title: 'network' },
  ransomware: { icon: LockKeyhole, tone: 'text-amber-600 bg-amber-500/10 ring-amber-500/20', title: 'ransomware' },
  script: { icon: Code2, tone: 'text-violet-500 bg-violet-500/10 ring-violet-500/20', title: 'script' },
  process: { icon: Workflow, tone: 'text-emerald-500 bg-emerald-500/10 ring-emerald-500/20', title: 'process' },
  endpoint: { icon: Bug, tone: 'text-orange-500 bg-orange-500/10 ring-orange-500/20', title: 'endpoint' },
}

function DetailRow({ label, value }: { label: string; value: string | null }) {
  if (!value) return null
  return (
    <div className="grid grid-cols-[minmax(140px,0.75fr)_minmax(0,1.5fr)] items-start gap-4 border-b border-border/50 px-4 py-2.5 last:border-b-0">
      <dt className="text-[11px] text-muted-foreground">{label}</dt>
      <dd className="break-words font-mono text-xs text-foreground/90">{value}</dd>
    </div>
  )
}

export function isEdrDocument(flat: Record<string, unknown>): boolean {
  return isEdrEvent(flat)
}

export function EdrDetailPanel({ flat, t }: EdrDetailPanelProps) {
  const kind = detectKind(flat)
  const meta = kindMeta[kind]
  const Icon = meta.icon
  const labels = (name: string) => t(`logExplorer.edr.fields.${name}`)
  const common = [
    ['host', text(flat, ['target.host', 'host.name', 'device.name', 'deviceName', 'computer'])],
    ['action', text(flat, ['action', 'event.action', 'activity'])],
    ['severity', text(flat, ['severity', 'event.severity', 'risk'])],
  ] as const
  const details = kind === 'file'
    ? [
        ['filePath', text(flat, ['file.path', 'filePath', 'path', 'target.file.path', 'target.path'])],
        ['verdict', text(flat, ['verdict', 'file.verdict', 'event.outcome', 'actionResult', 'status'])],
        ['detectionName', text(flat, ['detection.name', 'detectionName', 'malware.name', 'threat.name', 'signature'])],
        ['fileHash', text(flat, ['file.hash', 'file.hash.sha256', 'hash.sha256', 'sha256', 'hash', 'md5'])],
        ['process', text(flat, ['process.name', 'process', 'process.executable'])],
        ['parentProcess', text(flat, ['process.parent.name', 'parent.process.name', 'parentProcess', 'process.parent'])],
      ]
    : kind === 'network'
      ? [
          ['remoteAddress', text(flat, ['network.remote.address', 'remote.address', 'destination.ip', 'target.ip', 'remoteIp', 'target.address'])],
          ['domain', text(flat, ['network.remote.domain', 'remote.domain', 'dns.question.name', 'target.domain', 'domain'])],
          ['matchedIndicator', text(flat, ['indicator.match', 'matched.indicator', 'threat.indicator', 'indicator', 'ioc'])],
        ]
      : kind === 'ransomware'
        ? [
            ['scoreSignal', text(flat, ['ransomware.score.signal', 'score.signal', 'scoreSignal', 'detection.signal', 'signal'])],
            ['contained', text(flat, ['ransomware.contained', 'contained', 'isContained', 'containment', 'actionResult'])],
          ]
        : kind === 'script'
          ? [
              ['script', text(flat, ['script.name', 'script', 'command', 'process.command_line', 'commandLine'])],
              ['verdict', text(flat, ['verdict', 'event.outcome', 'actionResult', 'status'])],
              ['process', text(flat, ['process.name', 'process', 'process.executable'])],
              ['parentProcess', text(flat, ['process.parent.name', 'parent.process.name', 'parentProcess'])],
            ]
          : kind === 'process'
            ? [
                ['process', text(flat, ['process.name', 'process.executable', 'process'])],
                ['parentProcess', text(flat, ['process.parent.name', 'parent.process.name', 'parentProcess', 'process.parent'])],
                ['commandLine', text(flat, ['process.command_line', 'commandLine', 'command'])],
                ['verdict', text(flat, ['verdict', 'event.outcome', 'actionResult', 'status'])],
              ]
            : [
                ['detectionName', text(flat, ['detection.name', 'detectionName', 'threat.name', 'signature', 'event.reason'])],
                ['verdict', text(flat, ['verdict', 'event.outcome', 'actionResult', 'status'])],
                ['process', text(flat, ['process.name', 'process.executable', 'process'])],
                ['filePath', text(flat, ['file.path', 'filePath', 'path'])],
              ]

  return (
    <section className="overflow-hidden rounded-md border border-border bg-card">
      <div className="flex items-center gap-3 border-b border-border/60 px-4 py-3">
        <div className={cn('flex h-8 w-8 items-center justify-center rounded-md ring-1', meta.tone)}>
          <Icon size={16} />
        </div>
        <div>
          <h3 className="text-sm font-medium">{t(`logExplorer.edr.sources.${meta.title}`)}</h3>
          <p className="text-[11px] text-muted-foreground">{t('logExplorer.edr.subtitle')}</p>
        </div>
      </div>
      <dl>
        {common.map(([key, value]) => <DetailRow key={key} label={labels(key)} value={value} />)}
        {details.map(([key, value]) => <DetailRow key={key} label={labels(String(key))} value={value} />)}
      </dl>
    </section>
  )
}
