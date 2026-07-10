import { ShieldOff } from 'lucide-react'

export function NotConfiguredState() {
  return (
    <div className="flex min-h-[60vh] items-center justify-center px-6">
      <div className="max-w-md rounded-2xl border border-white/10 bg-slate-900/50 p-8 text-center">
        <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-slate-800">
          <ShieldOff className="h-6 w-6 text-slate-400" />
        </div>
        <h2 className="text-lg font-semibold text-white">Threat Intelligence isn't configured</h2>
        <p className="mt-2 text-sm text-slate-400">
          Contact your administrator to connect an upstream Cyber Mantra / ThreatWinds instance for this environment.
        </p>
      </div>
    </div>
  )
}
