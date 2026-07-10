import { Sparkles } from 'lucide-react'

export function IntelInsightsCard() {
  // ponytail: static copy — swap to useTiChat prompt when product spec lands
  return (
    <div className="flex h-full flex-col gap-3 rounded-xl border border-border bg-card p-5">
      <div className="flex items-center gap-2">
        <Sparkles size={14} className="text-fuchsia-500" />
        <h3 className="text-sm font-semibold">SOC AI insights</h3>
      </div>

      <div className="rounded-lg border border-fuchsia-500/30 bg-gradient-to-br from-fuchsia-500/10 via-violet-500/5 to-transparent p-3 text-xs leading-relaxed">
        <span className="font-medium text-foreground">3 adversaries</span> in your env attributed to{' '}
        <span className="font-medium">Scattered Spider</span> — campaign appears active.{' '}
        <button className="text-primary hover:underline">View campaign</button>
      </div>

      <div className="rounded-lg border border-amber-500/30 bg-amber-500/5 p-3 text-xs leading-relaxed">
        <span className="font-medium text-foreground">2 KEV CVEs</span> applicable to your stack — no
        patches deployed yet. <button className="text-primary hover:underline">Review</button>
      </div>

      <div className="rounded-lg border border-red-500/30 bg-red-500/5 p-3 text-xs leading-relaxed">
        <span className="font-medium text-foreground">1 feed errored</span> · Recorded Future auth
        token expired 12h ago. <button className="text-primary hover:underline">Reconnect</button>
      </div>
    </div>
  )
}
