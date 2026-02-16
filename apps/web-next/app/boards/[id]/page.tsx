import Link from "next/link";
import { getBoardTopics } from "@/lib/api";

type Props = {
  params: Promise<{ id: string }>;
};

export default async function BoardPage({ params }: Props) {
  const { id } = await params;
  const data = await getBoardTopics(id);

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
    </main>
  );
}

