const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:9080";

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

async function request<T>(path: string): Promise<T> {
  const res = await fetch(`${API_BASE_URL}${path}`, {
    next: { revalidate: 15 }
  });
  if (!res.ok) {
    throw new Error(`API request failed: ${path}`);
  }
  const json = (await res.json()) as RespBody<T>;
  return json.data;
}

export async function getBoardTopics(boardID: string): Promise<{ list: Topic[]; total: number }> {
  return request<{ list: Topic[]; total: number }>(`/api/v1/boards/${boardID}/topics`);
}

export async function getTopicWiki(topicID: string): Promise<WikiRevision | null> {
  return request<WikiRevision | null>(`/api/v1/topics/${topicID}/wiki`);
}

export async function getContributors(topicID: string): Promise<Contributor[]> {
  return request<Contributor[]>(`/api/v1/topics/${topicID}/contributors`);
}

export async function getDocGraph(topicID: string): Promise<DocGraph> {
  return request<DocGraph>(`/api/v1/docs/graph?root_topic_id=${topicID}`);
}

