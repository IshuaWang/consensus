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

For create/reply actions, backend auth is required. Log in on Answer backend first
on the same host (prefer `127.0.0.1`) so cookies can be forwarded.

## Dev fallback

When backend is unavailable in development, the app falls back to empty data by default
so pages can still render.

```bash
# disable fallback and fail fast
export ENABLE_DEV_API_FALLBACK=0

# optional request timeout (ms)
export API_TIMEOUT_MS=3000
```
