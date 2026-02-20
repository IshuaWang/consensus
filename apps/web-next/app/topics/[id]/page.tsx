import Link from "next/link";
import { revalidatePath } from "next/cache";
import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import {
  ApiRequestError,
  applyMergeJob,
  createMergeJob,
  createTopicPost,
  createTopicWikiRevision,
  getCurrentUser,
  getContributors,
  getDocGraph,
  getMergeJob,
  getTopic,
  getTopicMergeJobs,
  getTopicPosts,
  getTopicWiki,
  getTopicWikiRevisions,
  setTopicSolution,
  votePost,
  voteTopic
} from "@/lib/api";
import { PostCreateForm, type PostCreateState } from "@/components/post-create-form";
import { ANSWER_TOKEN_COOKIE } from "@/lib/auth";

type Props = {
  params: Promise<{ id: string }>;
  searchParams: Promise<{
    notice?: string;
    error?: string;
    merge_job_id?: string;
    modal?: string;
    merge_post_id?: string;
  }>;
};

type TopicModal = "" | "wiki-history" | "publish-wiki" | "pending-jobs" | "author-merge";

type TopicRedirectOptions = {
  notice?: string;
  error?: string;
  mergeJobID?: string;
  modal?: TopicModal;
  mergePostID?: string;
};

function trimQueryMessage(raw: string | undefined): string {
  return raw ? raw.trim().slice(0, 200) : "";
}

function normalizeIDToken(raw: string | undefined): string {
  if (!raw) {
    return "";
  }
  return raw.trim().split(/[\s,]+/)[0] || "";
}

function parseIDList(raw: string): string[] {
  const ids = raw
    .split(/[\s,\n]+/)
    .map((item) => item.trim())
    .filter(Boolean);
  return Array.from(new Set(ids));
}

function parsePositiveInt(raw: string, fallback = 1): number {
  const n = Number(raw);
  if (!Number.isFinite(n) || n <= 0) {
    return fallback;
  }
  return Math.floor(n);
}

function formatReplyTime(raw: string): string {
  if (!raw) {
    return "Unknown time";
  }
  const date = new Date(raw);
  if (Number.isNaN(date.getTime())) {
    return raw;
  }
  return new Intl.DateTimeFormat("en-US", {
    dateStyle: "medium",
    timeStyle: "short"
  }).format(date);
}

function normalizeTopicModal(raw: string | undefined): TopicModal {
  switch ((raw || "").trim()) {
    case "wiki-history":
    case "publish-wiki":
    case "pending-jobs":
    case "author-merge":
      return raw as TopicModal;
    default:
      return "";
  }
}

function buildTopicActionRedirectURL(topicID: string, options: TopicRedirectOptions = {}): string {
  const qp = new URLSearchParams();
  if (options.notice) {
    qp.set("notice", options.notice.slice(0, 160));
  }
  if (options.error) {
    qp.set("error", options.error.slice(0, 160));
  }
  if (options.mergeJobID) {
    qp.set("merge_job_id", options.mergeJobID);
  }
  if (options.modal) {
    qp.set("modal", options.modal);
  }
  if (options.mergePostID) {
    qp.set("merge_post_id", options.mergePostID);
  }
  const suffix = qp.toString();
  if (!suffix) {
    return `/topics/${topicID}`;
  }
  return `/topics/${topicID}?${suffix}`;
}

export default async function TopicPage({ params, searchParams }: Props) {
  const { id } = await params;
  const query = await searchParams;
  const noticeMessage = trimQueryMessage(query.notice);
  const errorMessage = trimQueryMessage(query.error);
  const activeMergeJobID = normalizeIDToken(query.merge_job_id);
  const activeModal = normalizeTopicModal(query.modal);
  const activeMergePostID = normalizeIDToken(query.merge_post_id);

  const createPostAction = async (
    _state: PostCreateState,
    formData: FormData
  ): Promise<PostCreateState> => {
    "use server";
    const text = String(formData.get("original_text") ?? "").trim();
    if (!text) {
      return { error: "Reply content is required.", message: null };
    }
    const authToken = (await cookies()).get(ANSWER_TOKEN_COOKIE)?.value;
    if (!authToken) {
      return {
        error: `Post reply failed: login required. Open /login?from=${encodeURIComponent(`/topics/${id}`)}`,
        message: null
      };
    }
    try {
      await createTopicPost(id, { original_text: text }, { authToken });
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
      if (err instanceof Error) {
        return { error: `Post reply failed: ${err.message}`, message: null };
      }
      return { error: "Post reply failed: unknown error.", message: null };
    }
  };

  const voteTopicAction = async (formData: FormData): Promise<void> => {
    "use server";
    const raw = Number(formData.get("value"));
    const value = raw === -1 ? -1 : 1;
    const authToken = (await cookies()).get(ANSWER_TOKEN_COOKIE)?.value;
    if (!authToken) {
      redirect(
        buildTopicActionRedirectURL(id, {
          error: "Vote failed: please sign in first.",
          mergeJobID: activeMergeJobID
        })
      );
    }
    try {
      await voteTopic(id, { value }, { authToken });
    } catch (err) {
      const message = err instanceof Error ? err.message : "unknown error";
      redirect(
        buildTopicActionRedirectURL(id, {
          error: `Vote failed: ${message}`,
          mergeJobID: activeMergeJobID
        })
      );
    }
    revalidatePath(`/topics/${id}`);
    redirect(
      buildTopicActionRedirectURL(id, {
        notice: "Topic vote updated.",
        mergeJobID: activeMergeJobID
      })
    );
  };

  const votePostAction = async (formData: FormData): Promise<void> => {
    "use server";
    const postID = normalizeIDToken(String(formData.get("post_id") ?? ""));
    if (!postID) {
      redirect(
        buildTopicActionRedirectURL(id, {
          error: "Vote failed: missing reply id.",
          mergeJobID: activeMergeJobID
        })
      );
    }
    const raw = Number(formData.get("value"));
    const value = raw === -1 ? -1 : 1;
    const authToken = (await cookies()).get(ANSWER_TOKEN_COOKIE)?.value;
    if (!authToken) {
      redirect(
        buildTopicActionRedirectURL(id, {
          error: "Vote failed: please sign in first.",
          mergeJobID: activeMergeJobID
        })
      );
    }
    try {
      await votePost(postID, { value }, { authToken });
    } catch (err) {
      const message = err instanceof Error ? err.message : "unknown error";
      redirect(
        buildTopicActionRedirectURL(id, {
          error: `Vote failed: ${message}`,
          mergeJobID: activeMergeJobID
        })
      );
    }
    revalidatePath(`/topics/${id}`);
    redirect(
      buildTopicActionRedirectURL(id, {
        notice: "Reply vote updated.",
        mergeJobID: activeMergeJobID
      })
    );
  };

  const markSolvedAction = async (formData: FormData): Promise<void> => {
    "use server";
    const postID = normalizeIDToken(String(formData.get("post_id") ?? ""));
    if (!postID) {
      redirect(
        buildTopicActionRedirectURL(id, {
          error: "Mark solved failed: missing reply id.",
          mergeJobID: activeMergeJobID
        })
      );
    }
    const authToken = (await cookies()).get(ANSWER_TOKEN_COOKIE)?.value;
    if (!authToken) {
      redirect(
        buildTopicActionRedirectURL(id, {
          error: "Mark solved failed: please sign in first.",
          mergeJobID: activeMergeJobID
        })
      );
    }
    try {
      await setTopicSolution(id, { post_id: postID }, { authToken });
    } catch (err) {
      const message = err instanceof Error ? err.message : "unknown error";
      redirect(
        buildTopicActionRedirectURL(id, {
          error: `Mark solved failed: ${message}`,
          mergeJobID: activeMergeJobID
        })
      );
    }
    revalidatePath(`/topics/${id}`);
    redirect(
      buildTopicActionRedirectURL(id, {
        notice: "Solved state updated.",
        mergeJobID: activeMergeJobID
      })
    );
  };

  const publishWikiRevisionAction = async (formData: FormData): Promise<void> => {
    "use server";
    const authToken = (await cookies()).get(ANSWER_TOKEN_COOKIE)?.value;
    if (!authToken) {
      redirect(
        buildTopicActionRedirectURL(id, {
          error: "Publish revision failed: please sign in first.",
          mergeJobID: activeMergeJobID,
          modal: activeModal
        })
      );
    }
    const title = String(formData.get("title") ?? "").trim();
    const summary = String(formData.get("summary") ?? "").trim();
    const document = String(formData.get("document") ?? "").trim();
    const sourcePostIDs = parseIDList(String(formData.get("source_post_ids") ?? ""));
    if (!title || !document) {
      redirect(
        buildTopicActionRedirectURL(id, {
          error: "Publish revision failed: title and document are required.",
          mergeJobID: activeMergeJobID,
          modal: activeModal
        })
      );
    }
    try {
      await createTopicWikiRevision(
        id,
        {
          title,
          summary,
          document,
          source_post_ids: sourcePostIDs
        },
        { authToken }
      );
    } catch (err) {
      const message = err instanceof Error ? err.message : "unknown error";
      redirect(
        buildTopicActionRedirectURL(id, {
          error: `Publish revision failed: ${message}`,
          mergeJobID: activeMergeJobID,
          modal: activeModal
        })
      );
    }
    revalidatePath(`/topics/${id}`);
    redirect(
      buildTopicActionRedirectURL(id, {
        notice: "New wiki revision published.",
        mergeJobID: activeMergeJobID
      })
    );
  };

  const createMergeJobAction = async (formData: FormData): Promise<void> => {
    "use server";
    const authToken = (await cookies()).get(ANSWER_TOKEN_COOKIE)?.value;
    if (!authToken) {
      redirect(
        buildTopicActionRedirectURL(id, {
          error: "Create merge job failed: please sign in first.",
          mergeJobID: activeMergeJobID
        })
      );
    }
    const postIDs = parseIDList(String(formData.get("post_ids") ?? ""));
    const summary = String(formData.get("summary") ?? "").trim();
    const nextModal = normalizeTopicModal(String(formData.get("next_modal") ?? ""));
    if (postIDs.length === 0) {
      redirect(
        buildTopicActionRedirectURL(id, {
          error: "Create merge job failed: add at least one reply id.",
          mergeJobID: activeMergeJobID,
          modal: activeModal,
          mergePostID: activeMergePostID
        })
      );
    }
    try {
      const job = await createMergeJob(
        id,
        {
          post_ids: postIDs,
          summary
        },
        { authToken }
      );
      revalidatePath(`/topics/${id}`);
      redirect(
        buildTopicActionRedirectURL(id, {
          notice: `Merge job ${job.id} created.`,
          mergeJobID: job.id,
          modal: nextModal || "pending-jobs"
        })
      );
    } catch (err) {
      const message = err instanceof Error ? err.message : "unknown error";
      redirect(
        buildTopicActionRedirectURL(id, {
          error: `Create merge job failed: ${message}`,
          mergeJobID: activeMergeJobID,
          modal: activeModal,
          mergePostID: activeMergePostID
        })
      );
    }
  };

  const applyMergeJobAction = async (formData: FormData): Promise<void> => {
    "use server";
    const authToken = (await cookies()).get(ANSWER_TOKEN_COOKIE)?.value;
    if (!authToken) {
      redirect(
        buildTopicActionRedirectURL(id, {
          error: "Apply merge job failed: please sign in first.",
          mergeJobID: activeMergeJobID,
          modal: "pending-jobs"
        })
      );
    }
    const mergeJobID = normalizeIDToken(String(formData.get("merge_job_id") ?? ""));
    const title = String(formData.get("title") ?? "").trim();
    const summary = String(formData.get("summary") ?? "").trim();
    const document = String(formData.get("document") ?? "").trim();
    const contributionWeight = parsePositiveInt(String(formData.get("contribution_weight") ?? ""), 1);
    if (!mergeJobID) {
      redirect(
        buildTopicActionRedirectURL(id, {
          error: "Apply merge job failed: missing merge job id.",
          mergeJobID: activeMergeJobID,
          modal: "pending-jobs"
        })
      );
    }
    if (!title || !document) {
      redirect(
        buildTopicActionRedirectURL(id, {
          error: "Apply merge job failed: title and document are required.",
          mergeJobID: mergeJobID || activeMergeJobID,
          modal: "pending-jobs"
        })
      );
    }
    try {
      await applyMergeJob(
        id,
        mergeJobID,
        {
          title,
          summary,
          document,
          contribution_weight: contributionWeight
        },
        { authToken }
      );
    } catch (err) {
      const message = err instanceof Error ? err.message : "unknown error";
      redirect(
        buildTopicActionRedirectURL(id, {
          error: `Apply merge job failed: ${message}`,
          mergeJobID,
          modal: "pending-jobs"
        })
      );
    }
    revalidatePath(`/topics/${id}`);
    redirect(
      buildTopicActionRedirectURL(id, {
        notice: `Merge job ${mergeJobID} applied and replies archived.`,
        mergeJobID,
        modal: "pending-jobs"
      })
    );
  };

  const mergeReplyAction = async (formData: FormData): Promise<void> => {
    "use server";
    const authToken = (await cookies()).get(ANSWER_TOKEN_COOKIE)?.value;
    if (!authToken) {
      redirect(
        buildTopicActionRedirectURL(id, {
          error: "Merge failed: please sign in first.",
          mergeJobID: activeMergeJobID,
          modal: activeModal
        })
      );
    }

    const postID = normalizeIDToken(String(formData.get("post_id") ?? ""));
    if (!postID) {
      redirect(
        buildTopicActionRedirectURL(id, {
          error: "Merge failed: missing reply id.",
          mergeJobID: activeMergeJobID,
          modal: activeModal
        })
      );
    }

    let targetJobID = "";
    try {
      const [topicSnapshot, wikiSnapshot, postsSnapshot, currentUser] = await Promise.all([
        getTopic(id, { authToken }),
        getTopicWiki(id, { authToken }),
        getTopicPosts(id, { authToken }),
        getCurrentUser({ authToken })
      ]);
      if (!topicSnapshot) {
        throw new Error("topic not found.");
      }
      if (!currentUser?.id) {
        throw new Error("login required.");
      }

      const targetPost = postsSnapshot.list.find((item) => item.id === postID);
      if (!targetPost) {
        throw new Error("reply not found.");
      }
      if (targetPost.merge_state === "archived") {
        redirect(
          buildTopicActionRedirectURL(id, {
            notice: `Reply ${postID} is already merged.`,
            mergeJobID: activeMergeJobID,
            modal: activeModal
          })
        );
      }

      const canQuickMerge = currentUser.is_admin_moderator || currentUser.id === topicSnapshot.user_id;
      if (!canQuickMerge) {
        if (currentUser.id === targetPost.user_id) {
          redirect(
            buildTopicActionRedirectURL(id, {
              notice: `Reply ${postID} prepared for merge proposal.`,
              mergeJobID: activeMergeJobID,
              modal: "author-merge",
              mergePostID: postID
            })
          );
        }
        redirect(
          buildTopicActionRedirectURL(id, {
            error: "Merge failed: only moderators/topic wiki editors can quick-merge.",
            mergeJobID: activeMergeJobID,
            modal: activeModal
          })
        );
      }

      const quickJob = await createMergeJob(
        id,
        {
          post_ids: [postID],
          summary: `Quick merge reply ${postID}`
        },
        { authToken }
      );
      targetJobID = quickJob.id;

      const revisionTitle = (wikiSnapshot?.title || topicSnapshot.title || `Topic ${id}`).trim();
      const revisionDocument = (
        wikiSnapshot?.document ||
        targetPost.parsed_text ||
        targetPost.original_text ||
        `Merged from reply ${postID}`
      ).trim();

      await applyMergeJob(
        id,
        quickJob.id,
        {
          title: revisionTitle || `Topic ${id}`,
          summary: `Quick merged reply ${postID}`,
          document: revisionDocument,
          contribution_weight: 1
        },
        { authToken }
      );
    } catch (err) {
      let message = "unknown error";
      if (err instanceof ApiRequestError && (err.status === 401 || err.status === 403)) {
        message = "moderator or topic wiki editor permission required.";
      } else if (err instanceof Error) {
        message = err.message;
      }
      redirect(
        buildTopicActionRedirectURL(id, {
          error: `Merge failed: ${message}`,
          mergeJobID: targetJobID || activeMergeJobID,
          modal: activeModal
        })
      );
    }

    revalidatePath(`/topics/${id}`);
    redirect(
      buildTopicActionRedirectURL(id, {
        notice: `Reply ${postID} merged.`,
        mergeJobID: targetJobID || activeMergeJobID,
        modal: "pending-jobs"
      })
    );
  };

  const token = (await cookies()).get(ANSWER_TOKEN_COOKIE)?.value;
  const canInteract = Boolean(token);
  const auth = token ? { authToken: token } : undefined;

  const [topic, wiki, wikiRevisions, contributors, graph, replies] = await Promise.all([
    getTopic(id, auth),
    getTopicWiki(id, auth),
    getTopicWikiRevisions(id, auth),
    getContributors(id, auth),
    getDocGraph(id, auth),
    getTopicPosts(id, auth)
  ]);

  let currentUser: Awaited<ReturnType<typeof getCurrentUser>> = null;
  let mergeJobs: Awaited<ReturnType<typeof getTopicMergeJobs>> = { list: [], total: 0 };
  if (auth) {
    const [currentUserResult, mergeJobsResult] = await Promise.allSettled([
      getCurrentUser(auth),
      getTopicMergeJobs(id, auth)
    ]);
    if (currentUserResult.status === "fulfilled") {
      currentUser = currentUserResult.value;
    }
    if (mergeJobsResult.status === "fulfilled") {
      mergeJobs = mergeJobsResult.value;
    }
  }

  let mergeJobError = "";
  let mergeJobDetail: Awaited<ReturnType<typeof getMergeJob>> = null;
  if (activeMergeJobID) {
    try {
      mergeJobDetail = await getMergeJob(id, activeMergeJobID, auth);
      if (!mergeJobDetail) {
        mergeJobError = `Merge job ${activeMergeJobID} not found for this topic.`;
      }
    } catch (err) {
      mergeJobError = err instanceof Error ? err.message : "Unknown merge job error.";
    }
  }

  const solvedPostID = topic?.solved_post_id && topic.solved_post_id !== "0" ? topic.solved_post_id : "";
  const hasSolved = solvedPostID !== "";
  const revisionByID = new Set(wikiRevisions.map((item) => item.id));
  const mergeRefPostIDs = new Set((mergeJobDetail?.post_refs ?? []).map((item) => item.post_id));
  const archivedReplyCount = replies.list.filter((reply) => reply.merge_state === "archived").length;
  const viewerID = currentUser?.id || "";
  const canQuickMerge = Boolean(
    currentUser && topic && (currentUser.is_admin_moderator || currentUser.id === topic.user_id)
  );
  const pendingMergeJobs = mergeJobs.list.filter((item) => item.status === "pending");
  const closeModalURL = buildTopicActionRedirectURL(id, { mergeJobID: activeMergeJobID });

  return (
    <>
      <main className="shell topic-shell topic-layout">
        <aside className="panel topic-sidebar">
          <p className="hero-kicker">Catalog</p>
          <h3>Topic Outline</h3>
          <nav className="toc-nav">
            <a href="#wiki-top">First Post / Wiki</a>
            <a href="#replies">Replies</a>
            <a href="#contributors">Contributors</a>
            <a href="#docs-graph">Docs Graph</a>
          </nav>
          <div className="topic-chip-row topic-sidebar-meta">
            <span className="chip">revisions {wikiRevisions.length}</span>
            <span className="chip">replies {replies.total}</span>
            <span className="chip">pending merge {pendingMergeJobs.length}</span>
          </div>
          <div className="toc-actions">
            <Link
              className="toc-action"
              href={buildTopicActionRedirectURL(id, {
                mergeJobID: activeMergeJobID,
                modal: "wiki-history"
              })}
            >
              Wiki Revision History
            </Link>
            <Link
              className="toc-action"
              href={buildTopicActionRedirectURL(id, {
                mergeJobID: activeMergeJobID,
                modal: "publish-wiki"
              })}
            >
              Publish New Revision
            </Link>
            <Link
              className="toc-action"
              href={buildTopicActionRedirectURL(id, {
                mergeJobID: activeMergeJobID,
                modal: "pending-jobs"
              })}
            >
              Pending Merge Jobs
            </Link>
          </div>
        </aside>

        <section className="topic-main">
          <section className="page-head">
            <p className="hero-kicker">Topic</p>
            <h1>{topic?.title ?? wiki?.title ?? id}</h1>
            <p>Current revision: {wiki?.id ?? "none"}</p>
            {noticeMessage && <p className="form-success">{noticeMessage}</p>}
            {errorMessage && <p className="form-error">{errorMessage}</p>}
            <div className="topic-meta-row">
              <div className="topic-chip-row">
                {hasSolved && <span className="chip chip-solved">solved</span>}
                <span className="chip">topic votes {topic?.vote_count ?? 0}</span>
                <span className="chip">wiki editor {topic?.user_id || "-"}</span>
              </div>
              {canInteract ? (
                <div className="vote-actions">
                  <form action={voteTopicAction}>
                    <input name="value" type="hidden" value="1" />
                    <button type="submit">▲ Topic</button>
                  </form>
                  <form action={voteTopicAction}>
                    <input name="value" type="hidden" value="-1" />
                    <button type="submit">▼ Topic</button>
                  </form>
                </div>
              ) : (
                <p className="action-hint">Sign in at /login to vote and merge.</p>
              )}
            </div>
          </section>

          <article className="panel panel-main wiki-main" id="wiki-top">
            <header className="reply-panel-head">
              <h2>First Post · Canonical Wiki</h2>
              <div className="wiki-toolbar">
                <Link
                  href={buildTopicActionRedirectURL(id, {
                    mergeJobID: activeMergeJobID,
                    modal: "wiki-history"
                  })}
                >
                  Revision History
                </Link>
                <Link
                  href={buildTopicActionRedirectURL(id, {
                    mergeJobID: activeMergeJobID,
                    modal: "publish-wiki"
                  })}
                >
                  Publish Revision
                </Link>
                <Link
                  href={buildTopicActionRedirectURL(id, {
                    mergeJobID: activeMergeJobID,
                    modal: "pending-jobs"
                  })}
                >
                  Merge Jobs
                </Link>
              </div>
            </header>
            {wiki ? (
              <>
                <p className="summary">{wiki.summary || "No summary yet."}</p>
                <pre>{wiki.document}</pre>
              </>
            ) : (
              <p className="empty">No wiki revision has been published.</p>
            )}
          </article>

          <section className="panel reply-panel" id="replies">
            <header className="reply-panel-head">
              <h2>Replies</h2>
              <p>
                {replies.total} total · {archivedReplyCount} archived
              </p>
            </header>
            {replies.list.length === 0 ? (
              <p className="empty">No replies yet.</p>
            ) : (
              <ul className="reply-list">
                {replies.list.map((reply) => {
                  const author = reply.author_display_name || reply.author_username || `user:${reply.user_id}`;
                  const handle = reply.author_username ? `@${reply.author_username}` : "";
                  const isSolved = solvedPostID !== "" && solvedPostID === reply.id;
                  const isArchived = reply.merge_state === "archived";
                  const inActiveMergeJob = mergeRefPostIDs.has(reply.id);
                  const isReplyAuthor = viewerID !== "" && viewerID === reply.user_id;
                  return (
                    <li
                      key={reply.id}
                      className={`reply-item${isArchived ? " is-merged" : ""}`}
                      id={`reply-${reply.id}`}
                    >
                      <header className="reply-item-head">
                        <div>
                          <strong>{author}</strong>
                          {handle && <span>{handle}</span>}
                        </div>
                        <time dateTime={reply.created_at}>{formatReplyTime(reply.created_at)}</time>
                      </header>
                      <pre>{reply.parsed_text || reply.original_text}</pre>
                      <div className="reply-actions">
                        <span className="chip">id {reply.id}</span>
                        <span className="chip">votes {reply.vote_count}</span>
                        {isArchived && <span className="chip chip-archived">archived</span>}
                        {isArchived && reply.archived_at && (
                          <span className="chip">at {formatReplyTime(reply.archived_at)}</span>
                        )}
                        {inActiveMergeJob && <span className="chip">in job {mergeJobDetail?.job.id}</span>}
                        {isSolved && <span className="chip chip-solved">best answer</span>}
                      </div>
                      {isArchived ? (
                        <p className="action-hint">This reply is archived by merge workflow and stays traceable.</p>
                      ) : (
                        <div className="reply-actions">
                          {canInteract ? (
                            <>
                              <form action={votePostAction}>
                                <input name="post_id" type="hidden" value={reply.id} />
                                <input name="value" type="hidden" value="1" />
                                <button type="submit">▲ Reply</button>
                              </form>
                              <form action={votePostAction}>
                                <input name="post_id" type="hidden" value={reply.id} />
                                <input name="value" type="hidden" value="-1" />
                                <button type="submit">▼ Reply</button>
                              </form>
                            </>
                          ) : (
                            <span className="action-hint">Sign in to vote.</span>
                          )}
                          {!isSolved &&
                            (canInteract ? (
                              <form action={markSolvedAction}>
                                <input name="post_id" type="hidden" value={reply.id} />
                                <button type="submit">Mark Solved</button>
                              </form>
                            ) : (
                              <span className="action-hint">Sign in to mark solved.</span>
                            ))}
                          {canInteract && (canQuickMerge || isReplyAuthor) ? (
                            <form action={mergeReplyAction}>
                              <input name="post_id" type="hidden" value={reply.id} />
                              <button type="submit">
                                {canQuickMerge ? "Merge to Wiki (Quick)" : "Propose Merge Job"}
                              </button>
                            </form>
                          ) : (
                            canInteract && <span className="action-hint">Only author can propose merge.</span>
                          )}
                        </div>
                      )}
                    </li>
                  );
                })}
              </ul>
            )}
          </section>

          <section className="topic-grid support-grid">
            <article className="panel" id="contributors">
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

            <article className="panel" id="docs-graph">
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
        </section>
      </main>

      {activeModal !== "" && (
        <div className="modal-overlay">
          <section className="modal-card">
            <header className="modal-head">
              <h3>
                {activeModal === "wiki-history" && "Wiki Revision History"}
                {activeModal === "publish-wiki" && "Publish New Revision"}
                {activeModal === "pending-jobs" && "Pending Merge Jobs"}
                {activeModal === "author-merge" && "Create Merge Job"}
              </h3>
              <Link className="modal-close" href={closeModalURL}>
                Close
              </Link>
            </header>

            {activeModal === "wiki-history" && (
              <>
                {wikiRevisions.length === 0 ? (
                  <p className="empty">No revisions yet.</p>
                ) : (
                  <ul className="revision-list">
                    {wikiRevisions.map((revision) => {
                      const isCurrent = wiki?.id === revision.id;
                      return (
                        <li key={revision.id} className={`revision-item${isCurrent ? " is-current" : ""}`}>
                          <div className="revision-head">
                            <strong>{revision.title}</strong>
                            <span>{formatReplyTime(revision.created_at)}</span>
                          </div>
                          <p>{revision.summary || "No summary."}</p>
                          <div className="topic-chip-row">
                            <span className="chip">revision {revision.id}</span>
                            <span className="chip">editor {revision.editor_id}</span>
                            {isCurrent && <span className="chip chip-solved">current</span>}
                            {!isCurrent && revisionByID.has(revision.parent_revision_id) && (
                              <span className="chip">parent {revision.parent_revision_id}</span>
                            )}
                          </div>
                        </li>
                      );
                    })}
                  </ul>
                )}
              </>
            )}

            {activeModal === "publish-wiki" && (
              <form className="compose-form inline-compose" action={publishWikiRevisionAction}>
                <label htmlFor="wiki-title">Title</label>
                <input
                  id="wiki-title"
                  name="title"
                  maxLength={180}
                  required
                  defaultValue={wiki?.title || topic?.title || ""}
                />
                <label htmlFor="wiki-summary">Summary</label>
                <input id="wiki-summary" name="summary" maxLength={500} placeholder="What changed in this revision?" />
                <label htmlFor="wiki-document">Document</label>
                <textarea
                  id="wiki-document"
                  name="document"
                  rows={10}
                  required
                  maxLength={200000}
                  defaultValue={wiki?.document || ""}
                />
                <label htmlFor="wiki-source-posts">Source reply IDs (optional)</label>
                <input
                  id="wiki-source-posts"
                  name="source_post_ids"
                  placeholder="101300...,101300... (comma or whitespace separated)"
                />
                <button type="submit">Publish Revision</button>
              </form>
            )}

            {activeModal === "pending-jobs" && (
              <div className="modal-stack">
                <form className="compose-form inline-compose" action={createMergeJobAction}>
                  <h4>Create Merge Job</h4>
                  <input name="next_modal" type="hidden" value="pending-jobs" />
                  <label htmlFor="merge-post-ids">Reply IDs</label>
                  <textarea
                    id="merge-post-ids"
                    name="post_ids"
                    rows={4}
                    required
                    placeholder="One or many IDs, split by comma/space/newline"
                  />
                  <label htmlFor="merge-summary">Summary</label>
                  <input id="merge-summary" name="summary" maxLength={500} placeholder="Why this merge batch exists" />
                  <button type="submit">Create Merge Job</button>
                </form>

                <article className="panel merge-job-card">
                  <h4>Pending Jobs</h4>
                  {pendingMergeJobs.length === 0 ? (
                    <p className="empty">No pending merge jobs.</p>
                  ) : (
                    <ul className="mini-list">
                      {pendingMergeJobs.map((job) => (
                        <li key={job.id}>
                          <span>
                            <strong>{job.id}</strong> · {job.summary || "No summary"}
                          </span>
                          <Link
                            href={buildTopicActionRedirectURL(id, {
                              mergeJobID: job.id,
                              modal: "pending-jobs"
                            })}
                          >
                            Review
                          </Link>
                        </li>
                      ))}
                    </ul>
                  )}
                </article>

                {mergeJobError && <p className="form-error">{mergeJobError}</p>}
                {mergeJobDetail && (
                  <article className="panel merge-job-card">
                    <div className="revision-head">
                      <strong>Job {mergeJobDetail.job.id}</strong>
                      <span>status {mergeJobDetail.job.status}</span>
                    </div>
                    <div className="topic-chip-row">
                      <span className="chip">creator {mergeJobDetail.job.creator_id}</span>
                      <span className="chip">reviewer {mergeJobDetail.job.reviewer_id || "-"}</span>
                      <span className="chip">
                        applied revision {mergeJobDetail.job.applied_revision_id || "pending"}
                      </span>
                    </div>
                    <ul className="mini-list">
                      {mergeJobDetail.post_refs.map((ref) => (
                        <li key={ref.id}>
                          <span>
                            <code>{ref.post_id}</code>
                          </span>
                          <Link href={`#reply-${ref.post_id}`}>jump</Link>
                        </li>
                      ))}
                    </ul>
                    <form className="compose-form inline-compose" action={applyMergeJobAction}>
                      <input name="merge_job_id" type="hidden" value={mergeJobDetail.job.id} />
                      <label htmlFor="apply-title">Revision title</label>
                      <input
                        id="apply-title"
                        name="title"
                        maxLength={180}
                        required
                        defaultValue={wiki?.title || topic?.title || ""}
                      />
                      <label htmlFor="apply-summary">Revision summary</label>
                      <input
                        id="apply-summary"
                        name="summary"
                        maxLength={500}
                        defaultValue={mergeJobDetail.job.summary || ""}
                      />
                      <label htmlFor="apply-document">Revision document</label>
                      <textarea
                        id="apply-document"
                        name="document"
                        rows={8}
                        maxLength={200000}
                        required
                        defaultValue={wiki?.document || ""}
                      />
                      <label htmlFor="apply-weight">Contribution weight</label>
                      <input id="apply-weight" name="contribution_weight" min={1} type="number" defaultValue={1} />
                      <button type="submit">Apply Merge Job</button>
                    </form>
                  </article>
                )}
              </div>
            )}

            {activeModal === "author-merge" && (
              <form className="compose-form inline-compose" action={createMergeJobAction}>
                <h4>Create Merge Job as Reply Author</h4>
                <p className="form-tip">This creates a pending merge job for moderators/topic wiki editors to review.</p>
                <input name="next_modal" type="hidden" value="pending-jobs" />
                <label htmlFor="author-merge-post-ids">Reply IDs</label>
                <textarea
                  id="author-merge-post-ids"
                  name="post_ids"
                  rows={4}
                  required
                  defaultValue={activeMergePostID}
                  placeholder="One or many IDs, split by comma/space/newline"
                />
                <label htmlFor="author-merge-summary">Summary</label>
                <input
                  id="author-merge-summary"
                  name="summary"
                  maxLength={500}
                  placeholder="What should be merged into wiki?"
                />
                <button type="submit">Create Pending Merge Job</button>
              </form>
            )}
          </section>
        </div>
      )}
    </>
  );
}
