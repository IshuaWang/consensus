# web-next

Next.js App Router + TypeScript frontend for the forum-first workflow.

## Run

```bash
cd apps/web-next
npm install
npm run dev
```

Default URL: `http://localhost:3001`

## Environment

Set API base URL if backend is not on default address:

```bash
export API_BASE_URL=http://localhost:9080
# Optional (also supported)
export NEXT_PUBLIC_API_BASE_URL=http://localhost:9080
```

The app currently reads forum APIs from `/api/v1`.

Current MVP pages:
- `GET /boards/:id`: topic list + create topic form
- `GET /topics/:id`: wiki/contributors/graph + add reply form
- `GET /login`: sign in for `web-next` write actions
- `GET /`: create board + jump to board

For create/reply actions, backend auth is required:
1. Open `http://localhost:3001/login`
2. Sign in with Answer account (example: `admin@example.com`)
3. Token is stored in `httpOnly` cookie for server actions
4. Create your first board on `/`, then create topics under that board

## Dev fallback

When backend is unavailable in development, the app falls back to empty data by default
so pages can still render.

```bash
# disable fallback and fail fast
export ENABLE_DEV_API_FALLBACK=0

# optional request timeout (ms)
export API_TIMEOUT_MS=3000
```
