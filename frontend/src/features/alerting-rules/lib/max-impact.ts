export function maxImpact(r: { confidentiality: number; integrity: number; availability: number }): number {
  return Math.max(r.confidentiality, r.integrity, r.availability)
}
