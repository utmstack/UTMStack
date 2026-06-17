import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './app/styles/index.css'
import { resolveFederationMode } from './shared/config/mode-bootstrap'

// Resolve the run mode (instance vs federation) from the backend BEFORE loading
// the app graph, so IS_FEDERATION (mode.ts) reads the right value everywhere.
// The app is dynamically imported afterwards so its modules evaluate with the
// flag already set on window.__UTM_FEDERATION__.
async function boot() {
  await resolveFederationMode()

  const [{ App }, { bootstrapPlatformLanguage }] = await Promise.all([
    import('./app/App'),
    import('./shared/i18n'),
  ])

  // Platform-default language for users who haven't chosen one in this browser;
  // a per-user lang_key still overrides it once authenticated.
  void bootstrapPlatformLanguage()

  createRoot(document.getElementById('root')!).render(
    <StrictMode>
      <App />
    </StrictMode>,
  )
}

void boot()
