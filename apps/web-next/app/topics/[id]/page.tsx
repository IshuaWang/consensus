import { getContributors, getDocGraph, getTopicWiki } from "@/lib/api";

type Props = {
  params: Promise<{ id: string }>;
};

export default async function TopicPage({ params }: Props) {
  const { id } = await params;
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
    </main>
  );
}

