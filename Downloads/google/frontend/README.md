# FinGuard Frontend

Next.js 14 (App Router) + TypeScript + Tailwind + shadcn-style primitives.

Two apps in one:
- `/analyst/*` — analyst console (live alerts, cases, risk, documents)
- `/customer/*` — customer dashboard (chat, transactions, disputes)

## Layout

```
src/
├── app/
│   ├── (auth)/            # login + callback (NextAuth-style)
│   ├── analyst/           # analyst-only pages
│   ├── customer/          # customer-only pages
│   └── api/               # BFF route handlers
├── components/
│   ├── ui/                # primitives (Button, Card, Table, ...)
│   ├── analyst/           # domain (AlertCard, RiskGauge, SHAP)
│   ├── customer/          # domain (ChatWindow, TransactionList)
│   └── layout/            # Shell, Sidebar, Topbar
├── lib/                   # api-client, auth, ws, types, format
├── hooks/                 # useAlerts, useRiskScore, useChat
└── styles/
```

## Local development

```bash
npm install
cp .env.example .env.local
npm run dev  # http://localhost:3000
```

## Build & run production image

```bash
npm run build
docker build -t finguard-frontend:dev -f Dockerfile .
docker run --rm -p 3000:3000 \
  -e FRAUD_SERVICE_URL=http://host.docker.internal:8080 \
  finguard-frontend:dev
```

## Linting / Type checking

```bash
npm run lint
npm run type-check
npm test
```
