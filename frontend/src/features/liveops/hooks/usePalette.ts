import { useThemeContext } from '@/app/providers'

export interface Palette {
  dark: boolean
  attacker: string
  external: string
  perimeter: string
  internal: string
  cloud: string
  ok: string
  muted: string
  alert: string
  /** Stolen keys, the impossible-travel arc, a successful sign-in. */
  highlight: string
  /** Error answers from a web server. */
  warn: string
  text: string
  textSoft: string
  textMuted: string
  textDim: string
  textFaint: string
  line: string
  /** The page behind the scene (--background): drawn around text so lines never cut through it. */
  bg: string
  /** Machines, labels and cards drawn on the stage (the app's card colour). */
  surface: string
  surfaceAlt: string
  alertText: string
  alertSoft: string
  /** Background of the alert marker and of a cut-off outside machine's badge. */
  alertWash: string
  okWash: string
  okText: string
}

const DARK: Palette = {
  dark: true,
  // Red, amber, blue, violet and green as the app shows them on dark (400-500 shades, primary blue).
  attacker: '#f87171',
  external: '#f87171',
  perimeter: '#f59e0b',
  internal: 'hsl(211 95% 60%)',
  cloud: '#a78bfa',
  ok: '#10b981',
  muted: 'hsl(220 10% 58%)',
  alert: '#ef4444',
  highlight: '#eab308',
  warn: '#f59e0b',
  // --foreground, --muted-foreground, --border, --card, --popover (dark).
  text: 'hsl(220 15% 93%)',
  textSoft: 'hsl(220 15% 85%)',
  textMuted: 'hsl(220 10% 58%)',
  textDim: 'hsl(220 10% 48%)',
  textFaint: 'hsl(222 14% 36%)',
  line: 'hsl(222 20% 24%)',
  bg: 'hsl(222 25% 10%)',
  surface: 'hsl(222 22% 14%)',
  surfaceAlt: 'hsl(222 20% 17%)',
  alertText: '#fca5a5',
  alertSoft: '#fca5a5',
  alertWash: 'hsl(0 45% 18%)',
  okWash: 'hsl(160 45% 14%)',
  okText: '#6ee7b7',
}

const LIGHT: Palette = {
  dark: false,
  // The same colours one step darker, as the app's pills and text use on light (600 shades).
  attacker: '#dc2626',
  external: '#dc2626',
  perimeter: '#d97706',
  internal: 'hsl(211 100% 47%)',
  cloud: '#7c3aed',
  ok: '#059669',
  muted: 'hsl(218 16% 56%)',
  alert: '#dc2626',
  highlight: '#ca8a04',
  warn: '#d97706',
  // --foreground, --muted-foreground, --border, --card (light).
  text: 'hsl(222 42% 10%)',
  textSoft: 'hsl(222 30% 22%)',
  textMuted: 'hsl(218 16% 44%)',
  textDim: 'hsl(218 14% 54%)',
  textFaint: 'hsl(216 16% 70%)',
  line: 'hsl(216 22% 82%)',
  bg: 'hsl(216 30% 95%)',
  surface: 'hsl(0 0% 100%)',
  surfaceAlt: 'hsl(216 30% 97%)',
  alertText: '#b91c1c',
  alertSoft: '#b91c1c',
  alertWash: 'hsl(0 86% 96%)',
  okWash: 'hsl(152 60% 94%)',
  okText: '#047857',
}

/** The colours for the app's current light or dark mode. */
export function usePalette(): Palette {
  const { theme } = useThemeContext()
  return theme === 'dark' ? DARK : LIGHT
}
