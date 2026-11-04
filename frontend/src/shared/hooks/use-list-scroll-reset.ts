import { useEffect, type RefObject } from 'react'

export function useListScrollReset(page: number, first: number, ref: RefObject<HTMLElement | null>) {
  useEffect(() => {
    if (page !== first) return
    let el: HTMLElement | null = ref.current
    while (el) {
      if (el.scrollTop > 0) {
        el.scrollTop = 0
        return
      }
      el = el.parentElement
    }
  }, [page, first, ref])
}
