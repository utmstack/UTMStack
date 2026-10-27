import { useEffect, useMemo, useState } from 'react'
import { alertsHttpService as svc } from '../services/alerts-http.service'
import type { Alert, AlertEventItem } from '../types/alert.types'

/** How many echoes are read to gather their events. The newest come first, so a
 *  very long thread shows its most recent activity rather than all of it. */
const ECHO_PAGE = 100

const eventTime = (e: AlertEventItem) => Date.parse(e['@timestamp'] || e.timestamp || e.deviceTime || '') || 0

export interface AlertThreadEvents {
  events: AlertEventItem[]
  /** Echoes of the alert, as the list reported them. */
  echoes: number
  /** Echoes whose events are included; fewer than `echoes` when the thread is longer than one page. */
  echoesRead: number
  loading: boolean
}

/**
 * The events behind an alert and, when it has echoes, behind each of them.
 * A single-event rule stores only the event that fired it, so for an alert that
 * keeps repeating the series is spread across the echoes — each one a separate
 * alert holding its own event. Fetched only while `enabled`, i.e. when the
 * Events tab is open.
 */
export function useAlertThreadEvents(alert: Alert, enabled: boolean): AlertThreadEvents {
  const echoes = alert.echoes ?? 0
  const [children, setChildren] = useState<Alert[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    setChildren([])
    if (!enabled || echoes === 0) return
    let cancelled = false
    setLoading(true)
    svc
      .echoes(alert.id, 1, ECHO_PAGE)
      .then(({ data }) => {
        if (!cancelled) setChildren(data ?? [])
      })
      .catch(() => {
        if (!cancelled) setChildren([])
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [alert.id, echoes, enabled])

  const events = useMemo(() => {
    const seen = new Set<string>()
    const out: AlertEventItem[] = []
    const add = (e?: AlertEventItem) => {
      if (!e) return
      if (e.id) {
        if (seen.has(e.id)) return
        seen.add(e.id)
      }
      out.push(e)
    }
    ;(alert.events ?? []).forEach(add)
    add(alert.lastEvent)
    children.forEach((c) => {
      ;(c.events ?? []).forEach(add)
      add(c.lastEvent)
    })
    return out.sort((x, y) => eventTime(y) - eventTime(x))
  }, [alert.events, alert.lastEvent, children])

  return { events, echoes, echoesRead: children.length, loading }
}
