import { ReactNode, createContext, useCallback, useContext, useEffect, useRef, useState } from 'react'
import { alertsHttpService as svc } from '../services/alerts-http.service'
import { alertNotificationHandler, NEW_ALERT_NOTIFICATION_INDEX } from '@/shared/lib/regular_alerts_count_handler'

export type AlertCountHandler = (newCount: number) => void

interface AlertCountContextValue {
  /** Total open alerts from the latest tick (null until first response). */
  count: number | null
  /** New alerts accumulated since the baseline (delta between ticks). */
  newCount: number
  registerHandler: (handler: AlertCountHandler,index?:string) => string
  removeHandler: (index: string) => void
}

const AlertCountContext = createContext<AlertCountContextValue | null>(null)

const POLL_MS = 10000

export function AlertCountProvider({ children }: { children: ReactNode }) {
  const [count, setCount] = useState<number | null>(null)
  const [newCount, setNewCount] = useState(0)
  const prevRef = useRef<number | null>(null)
  const handlersRef = useRef<Map<string, AlertCountHandler>>(new Map())
  const tick = useCallback(async () => {
    try {
      const next = await svc.countOpen()
      const prev = prevRef.current
      // Same value may arrive twice (StrictMode double-mount in dev); only
      // treat a strictly higher value as genuinely new alerts.
      if (prev !== null && next > prev) {
        const delta = next - prev
        setNewCount(delta)
        for (const handler of handlersRef.current.values()) handler(delta)
      } else if (prev !== null && next < prev) {
        // Count dropped (alerts resolved/removed) — reset the baseline so
        // it only ever accumulates upward.
        setNewCount(0)
      }
      if (prev !== next) prevRef.current = next
      setCount(next)
    } catch {
      /* keep the last value on transient failures */
    }
  }, [])

  useEffect(() => {
    void tick()
    const id = setInterval(() => void tick(), POLL_MS)
    return () => clearInterval(id)
  }, [tick])

  const registerHandler = useCallback((handler: AlertCountHandler, index?: string) => {
    if (!index) {
      index = Math.random().toString(36).slice(4)
    }
    handlersRef.current.set(index, handler)
    return index
  }, [])

  // Register the default toast handler once on initial load.
  useEffect(() => {
    registerHandler(alertNotificationHandler, NEW_ALERT_NOTIFICATION_INDEX)
  }, [registerHandler])

  const removeHandler = useCallback((index: string) => {
    handlersRef.current.delete(index)
  }, [])

  return (
    <AlertCountContext.Provider value={{ count, newCount,  registerHandler, removeHandler }}>
      {children}
    </AlertCountContext.Provider>
  )
}

export function useAlertCount(): AlertCountContextValue {
  const ctx = useContext(AlertCountContext)
  if (!ctx) throw new Error('useAlertCount must be used within AlertCountProvider')
  return ctx
}
