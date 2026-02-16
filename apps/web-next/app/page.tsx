import Link from "next/link";
import { BoardJump } from "@/components/board-jump";

export default function HomePage() {
  return (
    <main className="shell">
      <section className="hero">
        <p className="hero-kicker">Knowledge Is Edited, Not Frozen</p>
        <h1>Consensus Forum</h1>
        <p className="hero-copy">
          One topic, one evolving best answer. Discussion remains open, signal remains high.
        </p>
      </section>

      <section className="cards">
        <article className="card card-accent">
          <h2>Forum-First</h2>
          <p>
            Topics and replies run like a forum, while the lead post behaves like an evolving wiki.
          </p>
        </article>
        <article className="card">
          <h2>Merge Workflow</h2>
          <p>
            Moderators merge discussion replies into canonical revisions and archive absorbed replies.
          </p>
        </article>
        <article className="card">
          <h2>Docs Graph</h2>
          <p>Related topics are linked as a knowledge graph, not isolated threads.</p>
        </article>
      </section>

      <section className="entry">
        <BoardJump />
        <Link className="entry-link" href="/boards/demo-board-id">
          Open Demo Board
        </Link>
      </section>
    </main>
  );
}

