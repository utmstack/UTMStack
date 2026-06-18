# UTMStack Frontend (React)

React rewrite of the legacy `frontend-legacy/` (Angular) UI.

## Stack

- React 19 + TypeScript ~5.6
- Vite 6 + Tailwind v4 (CSS-first config in `src/app/styles/index.css`)
- react-router-dom v7
- Radix UI + shadcn-style local primitives in `src/shared/components/ui/`
- axios (minimal wrapper in `src/shared/lib/api-client.ts`)
- sonner (toasts), lucide-react (icons)
- Vitest + jsdom for tests

## Layout

```
frontend/
├── index.html
├── vite.config.ts / vitest.config.ts / tsconfig.json
├── nginx/default.conf      # nginx config used by Dockerfile
├── Dockerfile
└── src/
    ├── main.tsx            # entry
    ├── app/
    │   ├── App.tsx         # BrowserRouter + Providers + Routes
    │   ├── providers/      # context tree (Auth, Tooltip, Toaster…)
    │   ├── routes/         # route table
    │   └── styles/         # Tailwind + theme tokens
    ├── features/<name>/    # one folder per feature (vertical slice)
    │   ├── components/
    │   ├── pages/
    │   ├── services/       # context + http service per feature
    │   ├── types/
    │   └── index.ts        # barrel
    └── shared/
        ├── components/ui/  # button, tooltip, … (extend as needed)
        ├── layouts/        # DashboardLayout shell
        ├── lib/            # api-client, utils
        └── hooks/
```

`features/auth/` is the canonical example — copy it for new features.

## Run

```sh
npm install
npm run dev     # vite dev server
npm run build   # production bundle into dist/
npm run test    # vitest
```

Required env (set in `.env.local`):

| Var | Default |
|---|---|
| `VITE_API_URL` | `http://localhost:8080/api/v1` |
| `VITE_IAM_API_URL` | URL of IAM service for auth |

## Adding a feature

1. `mkdir -p src/features/<name>/{components,pages,services,types}`
2. Add `index.ts` barrel exporting public surface (provider, hook, page).
3. If the feature has app-wide state, register its provider in `src/app/providers/index.tsx`.
4. Add routes to `src/app/routes/index.tsx`.
