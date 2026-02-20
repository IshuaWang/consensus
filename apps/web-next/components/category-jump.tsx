"use client";

import { FormEvent, useState } from "react";
import { useRouter } from "next/navigation";

export function CategoryJump() {
  const [categoryID, setCategoryID] = useState("");
  const router = useRouter();

  const onSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!categoryID.trim()) {
      return;
    }
    router.push(`/categories/${categoryID.trim()}`);
  };

  return (
    <form className="category-jump" onSubmit={onSubmit}>
      <label htmlFor="category-id">Category ID</label>
      <div className="category-jump-row">
        <input
          id="category-id"
          name="category-id"
          value={categoryID}
          onChange={(event) => setCategoryID(event.target.value)}
          placeholder="10001111111111111"
        />
        <button type="submit">Open Category</button>
      </div>
    </form>
  );
}
