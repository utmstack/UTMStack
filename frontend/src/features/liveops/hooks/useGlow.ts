// The soft glow around lit-up shapes in a scene (filters defined in SceneDefs).
export function useGlow(strong = false) {
  return strong ? 'url(#lo-glow-strong)' : 'url(#lo-glow)'
}
