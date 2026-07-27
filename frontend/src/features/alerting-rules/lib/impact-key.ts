export function impactKey(n: number): 'high' | 'medium' | 'low' | 'none' {
  if (n >= 3) return 'high'
  if (n === 2) return 'medium'
  if (n === 1) return 'low'
  return 'none'
}
