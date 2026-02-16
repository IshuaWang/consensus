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

function isDevFallbackEnabled(): boolean {
  return process.env.NODE_ENV !== "production" && ENABLE_DEV_FALLBACK;
}

function buildFallbackMessage(path: string, err?: unknown): string {
  const reason = err instanceof Error ? err.message : "unknown error";
  return `[web-next] API unavailable, fallback enabled for ${path} (${reason})`;
}

async function request<T>(path: string, fallback: T): Promise<T> {
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), API_TIMEOUT_MS);
  try {
    const res = await fetch(`${API_BASE_URL}${path}`, {
      next: { revalidate: 15 },
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
  opts?: { cookieHeader?: string }
): Promise<TResp> {
  const headers: HeadersInit = {
    "Content-Type": "application/json"
  };
  if (opts?.cookieHeader) {
    headers.cookie = opts.cookieHeader;
  }
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

export async function getBoardTopics(boardID: string): Promise<{ list: Topic[]; total: number }> {
  return request<{ list: Topic[]; total: number }>(`/api/v1/boards/${boardID}/topics`, {
    list: [],
    total: 0
  });
}

export async function getTopicWiki(topicID: string): Promise<WikiRevision | null> {
  return request<WikiRevision | null>(`/api/v1/topics/${topicID}/wiki`, null);
}

export async function getContributors(topicID: string): Promise<Contributor[]> {
  return request<Contributor[]>(`/api/v1/topics/${topicID}/contributors`, []);
}

export async function getDocGraph(topicID: string): Promise<DocGraph> {
  return request<DocGraph>(`/api/v1/docs/graph?root_topic_id=${topicID}`, {
    nodes: [],
    edges: []
  });
}

export async function createTopic(
  payload: {
    board_id: string;
    title: string;
    topic_kind: "discussion" | "knowledge";
    is_wiki_enabled: boolean;
  },
  opts?: { cookieHeader?: string }
): Promise<Topic> {
  return postJSON<typeof payload, Topic>("/api/v1/topics", payload, opts);
}

export async function createTopicPost(
  topicID: string,
  payload: { original_text: string },
  opts?: { cookieHeader?: string }
): Promise<Post> {
  return postJSON<typeof payload, Post>(`/api/v1/topics/${topicID}/posts`, payload, opts);
}
