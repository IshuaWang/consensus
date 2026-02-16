"use client";

import { useActionState } from "react";

export type TopicCreateState = {
  error: string | null;
};

const initialState: TopicCreateState = {
  error: null
};

type Props = {
  action: (state: TopicCreateState, formData: FormData) => Promise<TopicCreateState>;
};

export function TopicCreateForm({ action }: Props) {
  const [state, formAction, pending] = useActionState(action, initialState);

  return (
    <form className="panel compose-form" action={formAction}>
      <h2>Create Topic</h2>
      <p className="form-tip">This writes to Answer forum APIs. You need a valid login session.</p>
      <label htmlFor="topic-title">Title</label>
      <input
        id="topic-title"
        name="title"
        required
        maxLength={180}
        placeholder="What should this topic answer?"
      />
      <div className="form-row">
        <label htmlFor="topic-kind">Kind</label>
        <select id="topic-kind" name="topic_kind" defaultValue="discussion">
          <option value="discussion">discussion</option>
          <option value="knowledge">knowledge</option>
        </select>
      </div>
      <label className="checkbox-row">
        <input name="is_wiki_enabled" type="checkbox" defaultChecked />
        Enable wiki mode for the first post
      </label>
      {state.error && <p className="form-error">{state.error}</p>}
      <button disabled={pending} type="submit">
        {pending ? "Creating..." : "Create Topic"}
      </button>
    </form>
  );
}
