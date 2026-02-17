import Link from "next/link";
import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { BoardJump } from "@/components/board-jump";
import { BoardCreateForm, type BoardCreateState } from "@/components/board-create-form";
import { ANSWER_TOKEN_COOKIE } from "@/lib/auth";
import { ApiRequestError, createBoard } from "@/lib/api";

export default async function HomePage() {
  const createBoardAction = async (
    _state: BoardCreateState,
    formData: FormData
  ): Promise<BoardCreateState> => {
    "use server";
    const slug = String(formData.get("slug") ?? "").trim();
    const name = String(formData.get("name") ?? "").trim();
    const description = String(formData.get("description") ?? "").trim();
    if (!slug || !name) {
      return { error: "Board slug and name are required." };
    }
    const authToken = (await cookies()).get(ANSWER_TOKEN_COOKIE)?.value;
    if (!authToken) {
      return {
        error: "Create board failed: login required. Open /login first."
      };
    }
    let boardID = "";
    try {
      const board = await createBoard(
        {
          slug,
          name,
          description
        },
        { authToken }
      );
      boardID = board.id;
    } catch (err) {
      if (err instanceof ApiRequestError) {
        if (err.status === 401 || err.status === 403) {
          return { error: "Create board failed: admin/moderator permission required." };
        }
        return { error: `Create board failed: ${err.message}` };
      }
      return { error: "Create board failed: unknown error." };
    }
    if (!boardID) {
      return { error: "Create board failed: invalid backend response (missing board id)." };
    }
    redirect(`/boards/${boardID}`);
  };

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
        <div className="entry-stack">
          <Link className="entry-link" href="/login">
            Sign In To Post
          </Link>
          <p className="entry-tip">
            `demo-board-id` is only a placeholder. Create a board first to get a real board ID.
          </p>
        </div>
      </section>

      <section className="compose-wrap">
        <BoardCreateForm action={createBoardAction} />
      </section>
    </main>
  );
}
