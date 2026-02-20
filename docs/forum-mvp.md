# Forum MVP Extension (Answer Fork)

This document describes the implemented MVP extension for forum-first knowledge synthesis.

## API Base Paths

- Legacy Answer APIs: `/answer/api/v1/*`
- New forum APIs: `/api/v1/*`

## Implemented Endpoints

### Forum Core

- `GET /api/v1/categories`
- `POST /api/v1/categories`
- `GET /api/v1/categories/{id}/topics`
- `GET /api/v1/topics/{id}`
- `POST /api/v1/topics`
- `POST /api/v1/topics/{id}/posts`
- `GET /api/v1/topics/{id}/posts`

### Wiki + Merge Workflow

- `GET /api/v1/topics/{id}/wiki`
- `POST /api/v1/topics/{id}/wiki/revisions`
- `GET /api/v1/topics/{id}/wiki/revisions`
- `POST /api/v1/topics/{id}/merge-jobs`
- `GET /api/v1/topics/{id}/merge-jobs/{jobId}`
- `POST /api/v1/topics/{id}/merge-jobs/{jobId}/apply`

### Contributors + Docs Graph

- `GET /api/v1/topics/{id}/contributors`
- `POST /api/v1/docs/links`
- `GET /api/v1/docs/graph?root_topic_id=...`

### Solved + Voting

- `POST /api/v1/topics/{id}/solution`
- `POST /api/v1/posts/{id}/votes`
- `POST /api/v1/topics/{id}/votes`

### Platform

- `GET /api/v1/platform/plugins`
- `POST /api/v1/platform/plugins/{id}/config`
- `GET /api/v1/platform/config`

## Core Domain Invariants

- A topic has only one `current_wiki_revision_id` at any given time.
- Merge job status transitions are one-way: `pending -> reviewed -> applied`.
- Applying the same merge job twice returns idempotent success if revision is already applied.
- Posts merged into wiki are archived (`merge_state=archived`, `archived_at` set).

## Data Model Additions

- `categories`
- `topics`
- `posts`
- `wiki_revisions`
- `merge_jobs`
- `merge_job_post_refs`
- `contribution_credits`
- `doc_links`
- `topic_votes`
- `post_votes`
- `topic_solutions`

Migration version added: `v1.9.0`.
