const cell = (value: unknown) => {
  const text = Array.isArray(value) ? value.join(', ') : value == null ? '' : String(value)
  return /[",\n]/.test(text) ? `"${text.replace(/"/g, '""')}"` : text
}

/** Saves `rows` under `header` as `<name>-<timestamp>.csv`. */
export function downloadCsv(name: string, header: string[], rows: unknown[][]) {
  const csv = [header.map(cell).join(','), ...rows.map((row) => row.map(cell).join(','))].join('\n')
  const url = URL.createObjectURL(new Blob([csv], { type: 'text/csv;charset=utf-8' }))
  const link = document.createElement('a')
  link.href = url
  link.download = `${name}-${new Date().toISOString().slice(0, 19)}.csv`
  document.body.appendChild(link)
  link.click()
  link.remove()
  URL.revokeObjectURL(url)
}
