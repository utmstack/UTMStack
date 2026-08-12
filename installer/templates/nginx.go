package templates

import "strings"

func NginxCustomBadGateway(brandName, logoSrc string) string {
	out := strings.ReplaceAll(nginxCustomBadGatewayTmpl, "{{BRAND}}", brandName)
	return strings.ReplaceAll(out, "{{LOGO}}", logoSrc)
}

const nginxCustomBadGatewayTmpl string = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>{{BRAND}} — Maintenance</title>
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap">
  <style>
    :root {
      color-scheme: light;
      --background: hsl(216 30% 95%);
      --foreground: hsl(222 42% 10%);
      --card: hsl(0 0% 100%);
      --muted-foreground: hsl(218 16% 44%);
      --primary: hsl(211 100% 47%);
      --border: hsl(216 22% 86%);
      --radius: 0.75rem;
      --shadow: 0 10px 30px -12px hsl(222 42% 10% / 0.18), 0 2px 6px -2px hsl(222 42% 10% / 0.08);
    }
    @media (prefers-color-scheme: dark) {
      :root {
        color-scheme: dark;
        --background: hsl(222 25% 10%);
        --foreground: hsl(220 15% 93%);
        --card: hsl(222 22% 14%);
        --muted-foreground: hsl(220 10% 58%);
        --primary: hsl(211 100% 50%);
        --border: hsl(222 20% 20%);
        --shadow: 0 10px 30px -12px hsl(0 0% 0% / 0.6), 0 2px 6px -2px hsl(0 0% 0% / 0.4);
      }
    }
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body {
      font-family: 'Inter', system-ui, -apple-system, sans-serif;
      background-color: var(--background);
      color: var(--foreground);
      display: flex;
      justify-content: center;
      align-items: center;
      min-height: 100vh;
      padding: 20px;
      -webkit-font-smoothing: antialiased;
    }
    .container {
      width: 100%;
      max-width: 480px;
      background-color: var(--card);
      border: 1px solid var(--border);
      border-radius: var(--radius);
      padding: 40px;
      text-align: center;
      box-shadow: var(--shadow);
      animation: fadeUp 0.5s cubic-bezier(0.22, 1, 0.36, 1) both;
    }
    .icon { margin-bottom: 24px; display: flex; justify-content: center; }
    .icon img { width: 64px; height: auto; object-fit: contain; }
    h1 {
      font-size: 24px;
      font-weight: 600;
      letter-spacing: -0.01em;
      margin-bottom: 12px;
    }
    p {
      font-size: 15px;
      line-height: 1.6;
      color: var(--muted-foreground);
      margin-bottom: 28px;
    }
    p strong { color: var(--foreground); font-weight: 600; }
    .dot {
      display: inline-flex;
      align-items: center;
      gap: 8px;
      font-size: 13px;
      font-weight: 500;
      color: var(--primary);
      padding: 6px 12px;
      border: 1px solid var(--border);
      border-radius: 999px;
    }
    .dot::before {
      content: "";
      width: 8px;
      height: 8px;
      border-radius: 999px;
      background: var(--primary);
      animation: pulse 1.6s ease-in-out infinite;
    }
    .footer {
      font-size: 12px;
      margin-top: 32px;
      color: var(--muted-foreground);
    }
    @keyframes fadeUp {
      from { opacity: 0; transform: translateY(14px); }
      to { opacity: 1; transform: translateY(0); }
    }
    @keyframes pulse {
      0%, 100% { opacity: 1; }
      50% { opacity: 0.35; }
    }
    @media (prefers-reduced-motion: reduce) {
      .container, .dot::before { animation: none; }
    }
  </style>
</head>
<body>
  <div class="container">
    <div class="icon">
      <img src="{{LOGO}}" alt="{{BRAND}}" />
    </div>
    <h1>We're getting things ready</h1>
    <p><strong>{{BRAND}}</strong> is currently under maintenance.<br>Please check back in a few minutes.</p>
    <span class="dot">Reconnecting</span>
    <div class="footer">&copy; {{BRAND}} 2025. All rights reserved.</div>
  </div>
</body>
</html>
`
