import { revalidatePath } from "next/cache";
import { cookies } from "next/headers";
import { getContributors, getDocGraph, getTopicWiki } from "@/lib/api";
import { ApiRequestError, createTopicPost } from "@/lib/api";
import { PostCreateForm, type PostCreateState } from "@/components/post-create-form";

type Props = {
  params: Promise<{ id: string }>;
};

export default async function TopicPage({ params }: Props) {
  const { id } = await params;
  const createPostAction = async (
    _state: PostCreateState,
    formData: FormData
  ): Promise<PostCreateState> => {
    "use server";
    const text = String(formData.get("original_text") ?? "").trim();
    if (!text) {
      return { error: "Reply content is required.", message: null };
    }
    try {
      await createTopicPost(
        id,
        { original_text: text },
        { cookieHeader: (await cookies()).toString() }
      );
      revalidatePath(`/topics/${id}`);
      return { error: null, message: "Reply submitted." };
    } catch (err) {
      if (err instanceof ApiRequestError) {
        if (err.status === 401 || err.status === 403) {
          return {
            error: "Post reply failed: login required in Answer.",
            message: null
          };
        }
        return { error: `Post reply failed: ${err.message}`, message: null };
      }
      return { error: "Post reply failed: unknown error.", message: null };
    }
  };

  const [wiki, contributors, graph] = await Promise.all([
    getTopicWiki(id),
    getContributors(id),
    getDocGraph(id)
  ]);

  return (
    <main className="shell topic-shell">
      <section className="page-head">
        <p className="hero-kicker">Topic</p>
        <h1>{wiki?.title ?? id}</h1>
        <p>Current revision: {wiki?.id ?? "none"}</p>
      </section>

      <section className="topic-grid">
        <article className="panel panel-main">
          <h2>Current Wiki</h2>
          {wiki ? (
            <>
              <p className="summary">{wiki.summary || "No summary yet."}</p>
              <pre>{wiki.document}</pre>
            </>
          ) : (
            <p className="empty">No wiki revision has been published.</p>
          )}
        </article>

        <article className="panel">
          <h2>Contributors</h2>
          {contributors.length === 0 && <p className="empty">No contributors tracked yet.</p>}
          <ul className="mini-list">
            {contributors.map((item) => (
              <li key={item.user_id}>
                <span>{item.user_id}</span>
                <span>{item.weight}</span>
              </li>
            ))}
          </ul>
        </article>

        <article className="panel">
          <h2>Docs Graph</h2>
          <p>
            nodes {graph.nodes.length} · edges {graph.edges.length}
          </p>
          <ul className="mini-list">
            {graph.edges.slice(0, 8).map((edge) => (
              <li key={edge.id}>
                <span>{edge.source_topic_id}</span>
                <span>{edge.target_topic_id}</span>
              </li>
            ))}
          </ul>
        </article>
      </section>

      <section className="compose-wrap">
        <PostCreateForm action={createPostAction} />
      </section>
    </main>
  );
}
