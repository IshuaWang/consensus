import Link from "next/link";
import { cookies } from "next/headers";
import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";
import { ApiRequestError, createTopic, getBoardTopics } from "@/lib/api";
import { TopicCreateForm, type TopicCreateState } from "@/components/topic-create-form";

type Props = {
  params: Promise<{ id: string }>;
};

export default async function BoardPage({ params }: Props) {
  const { id } = await params;
  const data = await getBoardTopics(id);
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
    try {
      const topic = await createTopic(
        {
          board_id: id,
          title,
          topic_kind: topicKind === "knowledge" ? "knowledge" : "discussion",
          is_wiki_enabled: wikiEnabled
        },
        { cookieHeader: (await cookies()).toString() }
      );
      revalidatePath(`/boards/${id}`);
      redirect(`/topics/${topic.id}`);
    } catch (err) {
      if (err instanceof ApiRequestError) {
        if (err.status === 401 || err.status === 403) {
          return { error: "Create topic failed: login required in Answer." };
        }
        return { error: `Create topic failed: ${err.message}` };
      }
      return { error: "Create topic failed: unknown error." };
    }
  };

  return (
    <main className="shell">
      <section className="page-head">
        <p className="hero-kicker">Board</p>
        <h1>{id}</h1>
        <p>{data.total} topics</p>
      </section>

      <section className="topic-list">
        {data.list.length === 0 && <p className="empty">No topics yet for this board.</p>}
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
