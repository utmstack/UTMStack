import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { App } from './app/App'
import { bootstrapPlatformLanguage } from './shared/i18n' // initializes i18next (side effect) before first render
import './app/styles/index.css'

// Pull the platform-default language for users who haven't chosen one in this
// browser; a per-user lang_key still overrides it once authenticated.
void bootstrapPlatformLanguage()

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
