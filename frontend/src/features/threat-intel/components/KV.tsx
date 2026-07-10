import React from 'react'

interface KVProps {
  k: string
  children: React.ReactNode
}

export function KV({ k, children }: KVProps) {
  return (
    <>
      <dt className="text-muted-foreground">{k}</dt>
      <dd className="break-words">{children}</dd>
    </>
  )
}
