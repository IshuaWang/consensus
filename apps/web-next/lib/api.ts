const API_BASE_URL =
  process.env.API_BASE_URL ?? process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:9080";
const API_TIMEOUT_MS = Number(process.env.API_TIMEOUT_MS ?? 3000);
const ENABLE_DEV_FALLBACK = process.env.ENABLE_DEV_API_FALLBACK !== "0";

type RespBody<T> = {
  code: number;
  reason: string;
  msg: string;
  data: T;
};

type RequestOptions = {
  authToken?: string;
  cookieHeader?: string;
};

type RawRecord = Record<string, unknown>;

export type Topic = {
  id: string;
  board_id: string;
  user_id: string;
  title: string;
  topic_kind: "discussion" | "knowledge";
  is_wiki_enabled: boolean;
  current_wiki_revision_id: string;
  solved_post_id: string;
  status: string;
  post_count: number;
  vote_count: number;
  last_post_id: string;
  created_at: string;
};

export type Board = {
  id: string;
  creator_id: string;
  slug: string;
  name: string;
  description: string;
  status: number;
  created_at: string;
};

export type Post = {
  id: string;
  topic_id: string;
  user_id: string;
  original_text: string;
  parsed_text: string;
  merge_state: string;
  archived_at?: string;
  vote_count: number;
  status: number;
  created_at: string;
};

export type WikiRevision = {
  id: string;
  topic_id: string;
  editor_id: string;
  title: string;
  document: string;
  summary: string;
  parent_revision_id: string;
  created_at: string;
};

export type Contributor = {
  user_id: string;
  weight: number;
};

export type DocGraph = {
  nodes: string[];
  edges: Array<{
    id: string;
    source_topic_id: string;
    target_topic_id: string;
    link_type: string;
  }>;
};

export type UserLogin = {
  access_token: string;
};

export class ApiRequestError extends Error {
  status: number;
  code?: number;
  reason?: string;

  constructor(message: string, status: number, code?: number, reason?: string) {
    super(message);
    this.name = "ApiRequestError";
    this.status = status;
    this.code = code;
    this.reason = reason;
  }
}

function asRecord(value: unknown): RawRecord {
  if (value && typeof value === "object" && !Array.isArray(value)) {
    return value as RawRecord;
  }
  return {};
}

function pick(source: RawRecord, ...keys: string[]): unknown {
  for (const key of keys) {
    if (key in source) {
      return source[key];
    }
  }
  return undefined;
}

function toStringSafe(value: unknown, fallback = ""): string {
  if (typeof value === "string") {
    return value;
  }
  if (typeof value === "number" || typeof value === "bigint") {
    return String(value);
  }
  return fallback;
}

function toNumberSafe(value: unknown, fallback = 0): number {
  if (typeof value === "number" && Number.isFinite(value)) {
    return value;
  }
  if (typeof value === "string") {
    const n = Number(value);
    if (Number.isFinite(n)) {
      return n;
    }
  }
  return fallback;
}

function toBoolSafe(value: unknown, fallback = false): boolean {
  if (typeof value === "boolean") {
    return value;
  }
  if (typeof value === "number") {
    return value !== 0;
  }
  if (typeof value === "string") {
    const v = value.trim().toLowerCase();
    if (v === "true" || v === "1") {
      return true;
    }
    if (v === "false" || v === "0") {
      return false;
    }
  }
  return fallback;
}

function mapBoard(raw: unknown): Board {
  const src = asRecord(raw);
  return {
    id: toStringSafe(pick(src, "id", "ID")),
    creator_id: toStringSafe(pick(src, "creator_id", "creatorID", "CreatorID")),
    slug: toStringSafe(pick(src, "slug", "Slug")),
    name: toStringSafe(pick(src, "name", "Name")),
    description: toStringSafe(pick(src, "description", "Description")),
    status: toNumberSafe(pick(src, "status", "Status")),
    created_at: toStringSafe(pick(src, "created_at", "createdAt", "CreatedAt"))
  };
}

function mapTopic(raw: unknown): Topic {
  const src = asRecord(raw);
  return {
    id: toStringSafe(pick(src, "id", "ID")),
    board_id: toStringSafe(pick(src, "board_id", "boardID", "BoardID")),
    user_id: toStringSafe(pick(src, "user_id", "userID", "UserID")),
    title: toStringSafe(pick(src, "title", "Title")),
    topic_kind:
      toStringSafe(pick(src, "topic_kind", "topicKind", "TopicKind")) === "knowledge"
        ? "knowledge"
        : "discussion",
    is_wiki_enabled: toBoolSafe(
      pick(src, "is_wiki_enabled", "isWikiEnabled", "IsWikiEnabled")
    ),
    current_wiki_revision_id: toStringSafe(
      pick(src, "current_wiki_revision_id", "currentWikiRevisionID", "CurrentWikiRevisionID")
    ),
    solved_post_id: toStringSafe(pick(src, "solved_post_id", "solvedPostID", "SolvedPostID")),
    status: toStringSafe(pick(src, "status", "Status")),
    post_count: toNumberSafe(pick(src, "post_count", "postCount", "PostCount")),
    vote_count: toNumberSafe(pick(src, "vote_count", "voteCount", "VoteCount")),
    last_post_id: toStringSafe(pick(src, "last_post_id", "lastPostID", "LastPostID")),
    created_at: toStringSafe(pick(src, "created_at", "createdAt", "CreatedAt"))
  };
}

function mapPost(raw: unknown): Post {
  const src = asRecord(raw);
  return {
    id: toStringSafe(pick(src, "id", "ID")),
    topic_id: toStringSafe(pick(src, "topic_id", "topicID", "TopicID")),
    user_id: toStringSafe(pick(src, "user_id", "userID", "UserID")),
    original_text: toStringSafe(pick(src, "original_text", "originalText", "Original")),
    parsed_text: toStringSafe(pick(src, "parsed_text", "parsedText", "Parsed")),
    merge_state: toStringSafe(pick(src, "merge_state", "mergeState", "MergeState")),
    archived_at: toStringSafe(pick(src, "archived_at", "archivedAt", "ArchivedAt")),
    vote_count: toNumberSafe(pick(src, "vote_count", "voteCount", "VoteCount")),
    status: toNumberSafe(pick(src, "status", "Status")),
    created_at: toStringSafe(pick(src, "created_at", "createdAt", "CreatedAt"))
  };
}

function mapWikiRevision(raw: unknown): WikiRevision {
  const src = asRecord(raw);
  return {
    id: toStringSafe(pick(src, "id", "ID")),
    topic_id: toStringSafe(pick(src, "topic_id", "topicID", "TopicID")),
    editor_id: toStringSafe(pick(src, "editor_id", "editorID", "EditorID")),
    title: toStringSafe(pick(src, "title", "Title")),
    document: toStringSafe(pick(src, "document", "Document")),
    summary: toStringSafe(pick(src, "summary", "Summary")),
    parent_revision_id: toStringSafe(
      pick(src, "parent_revision_id", "parentRevisionID", "ParentRevisionID")
    ),
    created_at: toStringSafe(pick(src, "created_at", "createdAt", "CreatedAt"))
  };
}

function mapContributor(raw: unknown): Contributor {
  const src = asRecord(raw);
  return {
    user_id: toStringSafe(pick(src, "user_id", "userID", "UserID")),
    weight: toNumberSafe(pick(src, "weight", "Weight"))
  };
}

function mapDocEdge(raw: unknown): DocGraph["edges"][number] {
  const src = asRecord(raw);
  return {
    id: toStringSafe(pick(src, "id", "ID")),
    source_topic_id: toStringSafe(
      pick(src, "source_topic_id", "sourceTopicID", "SourceTopicID")
    ),
    target_topic_id: toStringSafe(
      pick(src, "target_topic_id", "targetTopicID", "TargetTopicID")
    ),
    link_type: toStringSafe(pick(src, "link_type", "linkType", "LinkType"))
  };
}

function isDevFallbackEnabled(): boolean {
  return process.env.NODE_ENV !== "production" && ENABLE_DEV_FALLBACK;
}

function buildFallbackMessage(path: string, err?: unknown): string {
  const reason = err instanceof Error ? err.message : "unknown error";
  return `[web-next] API unavailable, fallback enabled for ${path} (${reason})`;
}

function applyAuthHeaders(headers: HeadersInit, opts?: RequestOptions): HeadersInit {
  const merged = { ...headers } as Record<string, string>;
  if (opts?.authToken) {
    merged.Authorization = `Bearer ${opts.authToken}`;
  }
  if (opts?.cookieHeader) {
    merged.cookie = opts.cookieHeader;
  }
  return merged;
}

async function request<T>(path: string, fallback: T, opts?: RequestOptions): Promise<T> {
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), API_TIMEOUT_MS);
  try {
    const res = await fetch(`${API_BASE_URL}${path}`, {
      next: { revalidate: 15 },
      headers: applyAuthHeaders({}, opts),
      signal: controller.signal
    });
    if (!res.ok) {
      if (isDevFallbackEnabled()) {
        console.warn(buildFallbackMessage(path, new Error(`http ${res.status}`)));
        return fallback;
      }
      throw new Error(`API request failed: ${path}`);
    }
    const json = (await res.json()) as RespBody<T>;
    return json.data;
  } catch (err) {
    if (isDevFallbackEnabled()) {
      console.warn(buildFallbackMessage(path, err));
      return fallback;
    }
    throw err;
  } finally {
    clearTimeout(timer);
  }
}

async function postJSON<TReq, TResp>(
  path: string,
  payload: TReq,
  opts?: RequestOptions
): Promise<TResp> {
  const headers = applyAuthHeaders(
    {
      "Content-Type": "application/json"
    },
    opts
  );
  const res = await fetch(`${API_BASE_URL}${path}`, {
    method: "POST",
    headers,
    body: JSON.stringify(payload),
    cache: "no-store"
  });
  let body: RespBody<TResp> | null = null;
  try {
    body = (await res.json()) as RespBody<TResp>;
  } catch {
    body = null;
  }
  if (!res.ok) {
    const message = body?.msg || `API request failed: ${path}`;
    throw new ApiRequestError(message, res.status, body?.code, body?.reason);
  }
  if (!body) {
    throw new ApiRequestError(`Empty response body: ${path}`, res.status);
  }
  return body.data;
}

export async function getBoardTopics(
  boardID: string,
  opts?: RequestOptions
): Promise<{ list: Topic[]; total: number }> {
  const raw = await request<{ list: unknown[]; total: number }>(
    `/api/v1/boards/${boardID}/topics`,
    {
      list: [],
      total: 0
    },
    opts
  );
  return {
    list: (raw.list ?? []).map(mapTopic),
    total: toNumberSafe(raw.total)
  };
}

export async function getTopicWiki(topicID: string, opts?: RequestOptions): Promise<WikiRevision | null> {
  const raw = await request<unknown | null>(`/api/v1/topics/${topicID}/wiki`, null, opts);
  if (!raw) {
    return null;
  }
  return mapWikiRevision(raw);
}

export async function getContributors(topicID: string, opts?: RequestOptions): Promise<Contributor[]> {
  const raw = await request<unknown[]>(`/api/v1/topics/${topicID}/contributors`, [], opts);
  return raw.map(mapContributor);
}

export async function getDocGraph(topicID: string, opts?: RequestOptions): Promise<DocGraph> {
  const raw = await request<{ nodes: unknown[]; edges: unknown[] }>(
    `/api/v1/docs/graph?root_topic_id=${topicID}`,
    {
      nodes: [],
      edges: []
    },
    opts
  );
  return {
    nodes: (raw.nodes ?? []).map((item) => toStringSafe(item)).filter(Boolean),
    edges: (raw.edges ?? []).map(mapDocEdge)
  };
}

export async function createTopic(
  payload: {
    board_id: string;
    title: string;
    topic_kind: "discussion" | "knowledge";
    is_wiki_enabled: boolean;
  },
  opts?: RequestOptions
): Promise<Topic> {
  const raw = await postJSON<typeof payload, unknown>("/api/v1/topics", payload, opts);
  return mapTopic(raw);
}

export async function createBoard(
  payload: {
    slug: string;
    name: string;
    description?: string;
  },
  opts?: RequestOptions
): Promise<Board> {
  const raw = await postJSON<typeof payload, unknown>("/api/v1/boards", payload, opts);
  return mapBoard(raw);
}

export async function createTopicPost(
  topicID: string,
  payload: { original_text: string },
  opts?: RequestOptions
): Promise<Post> {
  const raw = await postJSON<typeof payload, unknown>(`/api/v1/topics/${topicID}/posts`, payload, opts);
  return mapPost(raw);
}

export async function loginByEmail(payload: {
  e_mail: string;
  pass: string;
}): Promise<UserLogin> {
  return postJSON<typeof payload, UserLogin>("/answer/api/v1/user/login/email", payload);
}
