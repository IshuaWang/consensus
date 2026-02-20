const API_BASE_URL = normalizeAPIBaseURL(
  process.env.API_BASE_URL ?? process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://127.0.0.1:9080"
);
const API_TIMEOUT_MS = Number(process.env.API_TIMEOUT_MS ?? 8000);
const ENABLE_DEV_FALLBACK = process.env.ENABLE_DEV_API_FALLBACK !== "0";
const ENABLE_DEV_ENDPOINT_PROBING = process.env.ENABLE_DEV_ENDPOINT_PROBING === "1";
const DEV_API_BASE_URL_CANDIDATES = ENABLE_DEV_ENDPOINT_PROBING
  ? ["http://127.0.0.1:9080", "http://localhost:9080"]
  : [];

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

class ApiUnexpectedPayloadError extends Error {
  status: number;

  constructor(message: string, status: number) {
    super(message);
    this.name = "ApiUnexpectedPayloadError";
    this.status = status;
  }
}

export type Topic = {
  id: string;
  category_id: string;
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

export type Category = {
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

export type TopicPost = Post & {
  author_username: string;
  author_display_name: string;
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

export type MergeJob = {
  id: string;
  topic_id: string;
  creator_id: string;
  reviewer_id: string;
  status: string;
  summary: string;
  applied_revision_id: string;
  applied_at: string;
  created_at: string;
};

export type MergeJobPostRef = {
  id: string;
  merge_job_id: string;
  post_id: string;
  created_at: string;
};

export type MergeJobDetail = {
  job: MergeJob;
  post_refs: MergeJobPostRef[];
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

export type CurrentUser = {
  id: string;
  username: string;
  display_name: string;
  role_id: number;
  is_admin_moderator: boolean;
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

function normalizeAPIBaseURL(raw: string): string {
  return raw.replace(/\/+$/, "");
}

function dedupe(values: string[]): string[] {
  const set = new Set<string>();
  const result: string[] = [];
  for (const value of values) {
    if (!value || set.has(value)) {
      continue;
    }
    set.add(value);
    result.push(value);
  }
  return result;
}

function buildAPIBaseCandidates(): string[] {
  if (process.env.NODE_ENV === "production") {
    return [API_BASE_URL];
  }
  if (!ENABLE_DEV_ENDPOINT_PROBING) {
    return [API_BASE_URL];
  }
  return dedupe([API_BASE_URL, ...DEV_API_BASE_URL_CANDIDATES.map(normalizeAPIBaseURL)]);
}

function buildPathCandidates(path: string, withPrefixFallback = true): string[] {
  const candidates = [path];
  if (!withPrefixFallback || !ENABLE_DEV_ENDPOINT_PROBING) {
    return candidates;
  }
  if (path.startsWith("/api/v1")) {
    candidates.push(path.replace(/^\/api\/v1/, "/answer/api/v1"));
  }
  if (path.startsWith("/answer/api/v1")) {
    candidates.push(path.replace(/^\/answer\/api\/v1/, "/api/v1"));
  }
  return dedupe(candidates);
}

function buildEndpointCandidates(path: string, withPrefixFallback = true): string[] {
  const endpoints: string[] = [];
  const baseCandidates = buildAPIBaseCandidates();
  const pathCandidates = buildPathCandidates(path, withPrefixFallback);
  for (const base of baseCandidates) {
    for (const candidatePath of pathCandidates) {
      endpoints.push(`${base}${candidatePath}`);
    }
  }
  return dedupe(endpoints);
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

function mapCategory(raw: unknown): Category {
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
    category_id: toStringSafe(pick(src, "category_id", "categoryID", "CategoryID")),
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

function mapTopicPost(raw: unknown): TopicPost {
  const post = mapPost(raw);
  const src = asRecord(raw);
  return {
    ...post,
    author_username: toStringSafe(
      pick(src, "author_username", "authorUsername", "AuthorUsername", "username", "Username")
    ),
    author_display_name: toStringSafe(
      pick(
        src,
        "author_display_name",
        "authorDisplayName",
        "AuthorDisplayName",
        "display_name",
        "DisplayName"
      )
    )
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

function mapMergeJob(raw: unknown): MergeJob {
  const src = asRecord(raw);
  return {
    id: toStringSafe(pick(src, "id", "ID")),
    topic_id: toStringSafe(pick(src, "topic_id", "topicID", "TopicID")),
    creator_id: toStringSafe(pick(src, "creator_id", "creatorID", "CreatorID")),
    reviewer_id: toStringSafe(pick(src, "reviewer_id", "reviewerID", "ReviewerID")),
    status: toStringSafe(pick(src, "status", "Status")),
    summary: toStringSafe(pick(src, "summary", "Summary")),
    applied_revision_id: toStringSafe(
      pick(src, "applied_revision_id", "appliedRevisionID", "AppliedRevisionID")
    ),
    applied_at: toStringSafe(pick(src, "applied_at", "appliedAt", "AppliedAt")),
    created_at: toStringSafe(pick(src, "created_at", "createdAt", "CreatedAt"))
  };
}

function mapMergeJobPostRef(raw: unknown): MergeJobPostRef {
  const src = asRecord(raw);
  return {
    id: toStringSafe(pick(src, "id", "ID")),
    merge_job_id: toStringSafe(pick(src, "merge_job_id", "mergeJobID", "MergeJobID")),
    post_id: toStringSafe(pick(src, "post_id", "postID", "PostID")),
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

function mapCurrentUser(raw: unknown): CurrentUser {
  const src = asRecord(raw);
  const roleID = toNumberSafe(pick(src, "role_id", "roleID", "RoleID"));
  return {
    id: toStringSafe(pick(src, "id", "ID")),
    username: toStringSafe(pick(src, "username", "Username")),
    display_name: toStringSafe(pick(src, "display_name", "displayName", "DisplayName")),
    role_id: roleID,
    is_admin_moderator: roleID === 2 || roleID === 3
  };
}

function isDevFallbackEnabled(): boolean {
  return process.env.NODE_ENV !== "production" && ENABLE_DEV_FALLBACK;
}

function buildFallbackMessage(path: string, err?: unknown): string {
  const reason = err instanceof Error ? err.message : "unknown error";
  const endpoints = buildEndpointCandidates(path);
  return `[web-next] API unavailable, fallback enabled for ${path} (base=${API_BASE_URL}, tried=${endpoints.join(
    ", "
  )}, reason=${reason})`;
}

function summarizePayload(raw: string): string {
  return raw.replace(/\s+/g, " ").trim().slice(0, 120);
}

function isAbortError(err: unknown): boolean {
  return err instanceof Error && err.name === "AbortError";
}

function isHeadersTimeoutError(err: unknown): boolean {
  if (!(err instanceof TypeError)) {
    return false;
  }
  const cause = (err as { cause?: { code?: string } }).cause;
  return cause?.code === "UND_ERR_HEADERS_TIMEOUT";
}

function shouldTryNextCandidate(err: unknown): boolean {
  if (err instanceof ApiUnexpectedPayloadError) {
    return true;
  }
  if (err instanceof ApiRequestError) {
    return err.status === 404 || err.status === 405 || err.status === 502 || err.status === 503;
  }
  if (isAbortError(err)) {
    return true;
  }
  if (isHeadersTimeoutError(err)) {
    return true;
  }
  return err instanceof TypeError;
}

type NextFetchInit = RequestInit & {
  next?: {
    revalidate?: number;
  };
};

async function fetchWithTimeout(url: string, init: NextFetchInit): Promise<Response> {
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), API_TIMEOUT_MS);
  try {
    return await fetch(url, { ...init, signal: controller.signal });
  } finally {
    clearTimeout(timer);
  }
}

function normalizeFinalRequestError(path: string, err: unknown): Error {
  if (isAbortError(err) || isHeadersTimeoutError(err)) {
    return new ApiRequestError(`API timeout (${API_TIMEOUT_MS}ms): ${path}`, 504);
  }
  if (err instanceof Error) {
    return err;
  }
  return new Error(`API request failed: ${path}`);
}

async function parseJSONBody<T>(res: Response, path: string): Promise<RespBody<T>> {
  const contentType = (res.headers.get("content-type") || "").toLowerCase();
  if (!contentType.includes("application/json")) {
    const payload = summarizePayload(await res.text());
    throw new ApiUnexpectedPayloadError(
      `non-json response for ${path}: http ${res.status}, content-type=${
        contentType || "unknown"
      }, payload=${JSON.stringify(payload)}`,
      res.status
    );
  }
  try {
    return (await res.json()) as RespBody<T>;
  } catch (err) {
    throw new ApiUnexpectedPayloadError(
      `invalid json response for ${path}: ${err instanceof Error ? err.message : "unknown parse error"}`,
      res.status
    );
  }
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
  const endpoints = buildEndpointCandidates(path);
  let lastError: unknown = new Error(`API request failed: ${path}`);
  try {
    for (let i = 0; i < endpoints.length; i++) {
      const endpoint = endpoints[i];
      const canTryNext = i < endpoints.length - 1;
      try {
        const res = await fetchWithTimeout(endpoint, {
          next: { revalidate: 15 },
          headers: applyAuthHeaders({}, opts)
        });
        const json = await parseJSONBody<T>(res, endpoint);
        if (!res.ok) {
          const err = new ApiRequestError(
            json?.msg || `API request failed: ${path}`,
            res.status,
            json?.code,
            json?.reason
          );
          if (canTryNext && shouldTryNextCandidate(err)) {
            lastError = err;
            continue;
          }
          throw err;
        }
        return json.data;
      } catch (err) {
        if (canTryNext && shouldTryNextCandidate(err)) {
          lastError = err;
          continue;
        }
        throw normalizeFinalRequestError(path, err);
      }
    }
    throw normalizeFinalRequestError(path, lastError);
  } catch (err) {
    if (isDevFallbackEnabled()) {
      console.warn(buildFallbackMessage(path, err));
      return fallback;
    }
    throw normalizeFinalRequestError(path, err);
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
  // Write APIs in this project are fixed-prefix routes; avoid cross-prefix probing.
  const endpoints = buildEndpointCandidates(path, false);
  let lastError: unknown = new Error(`API request failed: ${path}`);
  for (let i = 0; i < endpoints.length; i++) {
    const endpoint = endpoints[i];
    const canTryNext = i < endpoints.length - 1;
    try {
      const res = await fetchWithTimeout(endpoint, {
        method: "POST",
        headers,
        body: JSON.stringify(payload),
        cache: "no-store"
      });
      const body = await parseJSONBody<TResp>(res, endpoint);
      if (!res.ok) {
        const err = new ApiRequestError(
          body?.msg || `API request failed: ${path}`,
          res.status,
          body?.code,
          body?.reason
        );
        if (canTryNext && shouldTryNextCandidate(err)) {
          lastError = err;
          continue;
        }
        throw err;
      }
      return body.data;
    } catch (err) {
      if (canTryNext && shouldTryNextCandidate(err)) {
        lastError = err;
        continue;
      }
      throw normalizeFinalRequestError(path, err);
    }
  }
  throw normalizeFinalRequestError(path, lastError);
}

export async function getCategoryTopics(
  categoryID: string,
  opts?: RequestOptions
): Promise<{ list: Topic[]; total: number }> {
  const raw = await request<{ list: unknown[]; total: number }>(
    `/api/v1/categories/${categoryID}/topics`,
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

export async function getCategories(
  opts?: RequestOptions
): Promise<{ list: Category[]; total: number }> {
  const raw = await request<{ list: unknown[]; total: number }>(
    "/api/v1/categories",
    {
      list: [],
      total: 0
    },
    opts
  );
  return {
    list: (raw.list ?? []).map(mapCategory),
    total: toNumberSafe(raw.total)
  };
}

export async function getTopicPosts(
  topicID: string,
  opts?: RequestOptions
): Promise<{ list: TopicPost[]; total: number }> {
  const raw = await request<{ list: unknown[]; total: number }>(
    `/api/v1/topics/${topicID}/posts`,
    {
      list: [],
      total: 0
    },
    opts
  );
  return {
    list: (raw.list ?? []).map(mapTopicPost),
    total: toNumberSafe(raw.total)
  };
}

export async function getTopic(topicID: string, opts?: RequestOptions): Promise<Topic | null> {
  const raw = await request<unknown | null>(`/api/v1/topics/${topicID}`, null, opts);
  if (!raw) {
    return null;
  }
  return mapTopic(raw);
}

export async function getTopicWiki(topicID: string, opts?: RequestOptions): Promise<WikiRevision | null> {
  const raw = await request<unknown | null>(`/api/v1/topics/${topicID}/wiki`, null, opts);
  if (!raw) {
    return null;
  }
  return mapWikiRevision(raw);
}

export async function getTopicWikiRevisions(
  topicID: string,
  opts?: RequestOptions
): Promise<WikiRevision[]> {
  const raw = await request<unknown[]>(`/api/v1/topics/${topicID}/wiki/revisions`, [], opts);
  return raw.map(mapWikiRevision);
}

export async function getMergeJob(
  topicID: string,
  jobID: string,
  opts?: RequestOptions
): Promise<MergeJobDetail | null> {
  const raw = await request<unknown | null>(`/api/v1/topics/${topicID}/merge-jobs/${jobID}`, null, opts);
  if (!raw) {
    return null;
  }
  const src = asRecord(raw);
  const jobRaw = pick(src, "job");
  if (!jobRaw) {
    return null;
  }
  const refsRaw = pick(src, "post_refs", "postRefs");
  const refs = Array.isArray(refsRaw) ? refsRaw : [];
  return {
    job: mapMergeJob(jobRaw),
    post_refs: refs.map(mapMergeJobPostRef)
  };
}

export async function getTopicMergeJobs(
  topicID: string,
  opts?: RequestOptions
): Promise<{ list: MergeJob[]; total: number }> {
  let raw: { list: unknown[]; total: number };
  try {
    raw = await request<{ list: unknown[]; total: number }>(
      `/api/v1/topics/${topicID}/merge-jobs`,
      {
        list: [],
        total: 0
      },
      opts
    );
  } catch {
    raw = await request<{ list: unknown[]; total: number }>(
      `/answer/api/v1/topics/${topicID}/merge-jobs`,
      {
        list: [],
        total: 0
      },
      opts
    );
  }
  return {
    list: (raw.list ?? []).map(mapMergeJob),
    total: toNumberSafe(raw.total)
  };
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
    category_id: string;
    title: string;
    topic_kind: "discussion" | "knowledge";
    is_wiki_enabled: boolean;
  },
  opts?: RequestOptions
): Promise<Topic> {
  const raw = await postJSON<typeof payload, unknown>("/api/v1/topics", payload, opts);
  return mapTopic(raw);
}

export async function createCategory(
  payload: {
    slug: string;
    name: string;
    description?: string;
  },
  opts?: RequestOptions
): Promise<Category> {
  const raw = await postJSON<typeof payload, unknown>("/api/v1/categories", payload, opts);
  return mapCategory(raw);
}

export async function createTopicPost(
  topicID: string,
  payload: { original_text: string },
  opts?: RequestOptions
): Promise<Post> {
  const raw = await postJSON<typeof payload, unknown>(`/api/v1/topics/${topicID}/posts`, payload, opts);
  return mapPost(raw);
}

export async function createTopicWikiRevision(
  topicID: string,
  payload: {
    title: string;
    document: string;
    summary?: string;
    source_post_ids?: string[];
  },
  opts?: RequestOptions
): Promise<WikiRevision> {
  const raw = await postJSON<typeof payload, unknown>(
    `/api/v1/topics/${topicID}/wiki/revisions`,
    payload,
    opts
  );
  return mapWikiRevision(raw);
}

export async function createMergeJob(
  topicID: string,
  payload: {
    post_ids: string[];
    summary?: string;
  },
  opts?: RequestOptions
): Promise<MergeJob> {
  const raw = await postJSON<typeof payload, unknown>(`/api/v1/topics/${topicID}/merge-jobs`, payload, opts);
  return mapMergeJob(raw);
}

export async function applyMergeJob(
  topicID: string,
  jobID: string,
  payload: {
    title: string;
    document: string;
    summary?: string;
    contribution_weight?: number;
  },
  opts?: RequestOptions
): Promise<WikiRevision> {
  const raw = await postJSON<typeof payload, unknown>(
    `/api/v1/topics/${topicID}/merge-jobs/${jobID}/apply`,
    payload,
    opts
  );
  return mapWikiRevision(raw);
}

export async function voteTopic(
  topicID: string,
  payload: { value: -1 | 1 },
  opts?: RequestOptions
): Promise<void> {
  await postJSON<typeof payload, unknown>(`/api/v1/topics/${topicID}/votes`, payload, opts);
}

export async function votePost(
  postID: string,
  payload: { value: -1 | 1 },
  opts?: RequestOptions
): Promise<void> {
  await postJSON<typeof payload, unknown>(`/api/v1/posts/${postID}/votes`, payload, opts);
}

export async function setTopicSolution(
  topicID: string,
  payload: { post_id: string },
  opts?: RequestOptions
): Promise<void> {
  await postJSON<typeof payload, unknown>(`/api/v1/topics/${topicID}/solution`, payload, opts);
}

export async function getCurrentUser(opts?: RequestOptions): Promise<CurrentUser | null> {
  let raw: unknown | null = null;
  try {
    raw = await request<unknown | null>("/api/v1/user/info", null, opts);
  } catch {
    raw = await request<unknown | null>("/answer/api/v1/user/info", null, opts);
  }
  if (!raw) {
    return null;
  }
  const user = mapCurrentUser(raw);
  if (!user.id) {
    return null;
  }
  return user;
}

export async function loginByEmail(payload: {
  e_mail: string;
  pass: string;
}): Promise<UserLogin> {
  return postJSON<typeof payload, UserLogin>("/answer/api/v1/user/login/email", payload);
}
