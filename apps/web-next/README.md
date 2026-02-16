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
export NEXT_PUBLIC_API_BASE_URL=http://localhost:9080
```

The app currently reads forum APIs from `/api/v1`.

