"use client";

import { useActionState } from "react";

export type BoardCreateState = {
  error: string | null;
};

const initialState: BoardCreateState = {
  error: null
};

type Props = {
  action: (state: BoardCreateState, formData: FormData) => Promise<BoardCreateState>;
};

export function BoardCreateForm({ action }: Props) {
  const [state, formAction, pending] = useActionState(action, initialState);

  return (
    <form className="panel compose-form" action={formAction}>
      <h2>Create Board</h2>
      <p className="form-tip">
        Create a real board first, then create topics under that board.
      </p>
      <label htmlFor="board-slug">Slug</label>
      <input
        id="board-slug"
        name="slug"
        required
        maxLength={100}
        placeholder="life-skills"
      />
      <label htmlFor="board-name">Name</label>
      <input
        id="board-name"
        name="name"
        required
        maxLength={120}
        placeholder="Life Skills"
      />
      <label htmlFor="board-description">Description</label>
      <textarea
        id="board-description"
        name="description"
        rows={3}
        maxLength={500}
        placeholder="Scope and posting rules for this board..."
      />
      {state.error && <p className="form-error">{state.error}</p>}
      <button disabled={pending} type="submit">
        {pending ? "Creating..." : "Create Board"}
      </button>
    </form>
  );
}

