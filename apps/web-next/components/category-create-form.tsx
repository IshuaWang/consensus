"use client";

import { useActionState } from "react";

export type CategoryCreateState = {
  error: string | null;
};

const initialState: CategoryCreateState = {
  error: null
};

type Props = {
  action: (state: CategoryCreateState, formData: FormData) => Promise<CategoryCreateState>;
};

export function CategoryCreateForm({ action }: Props) {
  const [state, formAction, pending] = useActionState(action, initialState);

  return (
    <form className="panel compose-form" action={formAction}>
      <h2>Create Category</h2>
      <p className="form-tip">
        Create a real category first, then create topics under that category.
      </p>
      <label htmlFor="category-slug">Category Slug</label>
      <input
        id="category-slug"
        name="slug"
        required
        maxLength={100}
        placeholder="life-skills"
      />
      <label htmlFor="category-name">Category Name</label>
      <input
        id="category-name"
        name="name"
        required
        maxLength={120}
        placeholder="Life Skills"
      />
      <label htmlFor="category-description">Description</label>
      <textarea
        id="category-description"
        name="description"
        rows={3}
        maxLength={500}
        placeholder="Scope and posting rules for this category..."
      />
      {state.error && <p className="form-error">{state.error}</p>}
      <button disabled={pending} type="submit">
        {pending ? "Creating..." : "Create Category"}
      </button>
    </form>
  );
}
