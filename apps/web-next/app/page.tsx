import Link from "next/link";
import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { CategoryCreateForm, type CategoryCreateState } from "@/components/category-create-form";
import { ANSWER_TOKEN_COOKIE } from "@/lib/auth";
import { ApiRequestError, createCategory, getCategories } from "@/lib/api";

export default async function HomePage() {
  const createCategoryAction = async (
    _state: CategoryCreateState,
    formData: FormData
  ): Promise<CategoryCreateState> => {
    "use server";
    const slug = String(formData.get("slug") ?? "").trim();
    const name = String(formData.get("name") ?? "").trim();
    const description = String(formData.get("description") ?? "").trim();
    if (!slug || !name) {
      return { error: "Category slug and name are required." };
    }
    const authToken = (await cookies()).get(ANSWER_TOKEN_COOKIE)?.value;
    if (!authToken) {
      return {
        error: "Create category failed: login required. Open /login first."
      };
    }
    let categoryID = "";
    try {
      const category = await createCategory(
        {
          slug,
          name,
          description
        },
        { authToken }
      );
      categoryID = category.id;
    } catch (err) {
      if (err instanceof ApiRequestError) {
        if (err.status === 401 || err.status === 403) {
          return { error: "Create category failed: admin/moderator permission required." };
        }
        return { error: `Create category failed: ${err.message}` };
      }
      return { error: "Create category failed: unknown error." };
    }
    if (!categoryID) {
      return { error: "Create category failed: invalid backend response (missing category id)." };
    }
    redirect(`/categories/${categoryID}`);
  };
  let categories: Awaited<ReturnType<typeof getCategories>>["list"] = [];
  let total = 0;
  let listError = "";
  try {
    const resp = await getCategories();
    categories = resp.list;
    total = resp.total;
  } catch (err) {
    if (err instanceof ApiRequestError) {
      listError = `Failed to load categories: ${err.message}`;
    } else {
      listError = "Failed to load categories: unknown error.";
    }
  }

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

      <section className="category-list-wrap panel">
        <header className="category-list-head">
          <h2>Categories</h2>
          <p>{total} total</p>
        </header>
        {listError ? (
          <p className="form-error">{listError}</p>
        ) : categories.length === 0 ? (
          <p className="empty">No categories yet. Create the first category below.</p>
        ) : (
          <ul className="category-list">
            {categories.map((category) => (
              <li key={category.id}>
                <Link href={`/categories/${category.id}`}>
                  <article className="topic-card">
                    <header>
                      <span className="chip">category</span>
                    </header>
                    <h3>{category.name}</h3>
                    <p>{category.description || "No description yet."}</p>
                    <p className="category-meta">
                      <code>{category.slug}</code> · <code>{category.id}</code>
                    </p>
                  </article>
                </Link>
              </li>
            ))}
          </ul>
        )}
      </section>

      <section className="entry entry-single">
        <div className="entry-stack">
          <Link className="entry-link" href="/login">
            Sign In To Post
          </Link>
          <p className="entry-tip">
            Topics can only be created inside a category. Pick one above or create a new category.
          </p>
        </div>
      </section>

      <section className="compose-wrap">
        <CategoryCreateForm action={createCategoryAction} />
      </section>
    </main>
  );
}
