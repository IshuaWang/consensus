"use client";

import Link from "next/link";
import { useActionState } from "react";

export type PostCreateState = {
  error: string | null;
  message: string | null;
};

const initialState: PostCreateState = {
  error: null,
  message: null
};

type Props = {
  action: (state: PostCreateState, formData: FormData) => Promise<PostCreateState>;
};

export function PostCreateForm({ action }: Props) {
  const [state, formAction, pending] = useActionState(action, initialState);

  return (
    <form className="panel compose-form" action={formAction}>
      <h2>Add Reply</h2>
      <p className="form-tip">
        Reply content can be merged into wiki revisions by moderators later. Sign in at{" "}
        <Link href="/login">/login</Link>.
      </p>
      <label htmlFor="post-content">Reply content</label>
      <textarea
        id="post-content"
        name="original_text"
        required
        maxLength={20000}
        placeholder="Write new evidence, corrections, or edge cases..."
        rows={7}
      />
      {state.error && <p className="form-error">{state.error}</p>}
      {state.message && <p className="form-success">{state.message}</p>}
      <button disabled={pending} type="submit">
        {pending ? "Posting..." : "Post Reply"}
      </button>
    </form>
  );
}
