import Link from "next/link";
import { cookies } from "next/headers";
import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";
import { ApiRequestError, createTopic, getCategoryTopics } from "@/lib/api";
import { TopicCreateForm, type TopicCreateState } from "@/components/topic-create-form";
import { ANSWER_TOKEN_COOKIE } from "@/lib/auth";

type Props = {
  params: Promise<{ id: string }>;
};

export default async function CategoryPage({ params }: Props) {
  const { id } = await params;
  const token = (await cookies()).get(ANSWER_TOKEN_COOKIE)?.value;
  const data = await getCategoryTopics(id, token ? { authToken: token } : undefined);
  const createTopicAction = async (
    _state: TopicCreateState,
    formData: FormData
  ): Promise<TopicCreateState> => {
    "use server";
    const title = String(formData.get("title") ?? "").trim();
    const topicKind = String(formData.get("topic_kind") ?? "discussion");
    const wikiEnabled = formData.get("is_wiki_enabled") === "on";
    if (!title) {
      return { error: "Topic title is required." };
    }
    const authToken = (await cookies()).get(ANSWER_TOKEN_COOKIE)?.value;
    if (!authToken) {
      return {
        error: `Create topic failed: login required. Open /login?from=${encodeURIComponent(`/categories/${id}`)}`
      };
    }
    let topicID = "";
    try {
      const topic = await createTopic(
        {
          category_id: id,
          title,
          topic_kind: topicKind === "knowledge" ? "knowledge" : "discussion",
          is_wiki_enabled: wikiEnabled
        },
        { authToken }
      );
      topicID = topic.id;
      revalidatePath(`/categories/${id}`);
    } catch (err) {
      if (err instanceof ApiRequestError) {
        if (err.status === 404 || /object not found/i.test(err.message)) {
          return {
            error: "Create topic failed: category not found. Create a category on the home page first."
          };
        }
        if (err.status === 401 || err.status === 403) {
          return { error: "Create topic failed: login required in Answer." };
        }
        return { error: `Create topic failed: ${err.message}` };
      }
      return { error: "Create topic failed: unknown error." };
    }
    if (!topicID) {
      return { error: "Create topic failed: invalid backend response (missing topic id)." };
    }
    redirect(`/topics/${topicID}`);
  };

  return (
    <main className="shell">
      <section className="page-head">
        <p className="hero-kicker">Category</p>
        <h1>{id}</h1>
        <p>{data.total} topics</p>
      </section>

      <section className="topic-list">
        {data.list.length === 0 && <p className="empty">No topics yet for this category.</p>}
        {data.list.map((topic) => (
          <article key={topic.id} className="topic-card">
            <header>
              <span className="chip">{topic.topic_kind}</span>
              <span className="chip">{topic.is_wiki_enabled ? "wiki" : "discussion"}</span>
            </header>
            <h2>{topic.title}</h2>
            <p>
              votes {topic.vote_count} · replies {topic.post_count}
            </p>
            <Link href={`/topics/${topic.id}`}>Open Topic</Link>
          </article>
        ))}
      </section>

      <section className="compose-wrap">
        <TopicCreateForm action={createTopicAction} />
      </section>
    </main>
  );
}
