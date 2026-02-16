"use client";

import { FormEvent, useState } from "react";
import { useRouter } from "next/navigation";

export function BoardJump() {
  const [boardID, setBoardID] = useState("");
  const router = useRouter();

  const onSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!boardID.trim()) {
      return;
    }
    router.push(`/boards/${boardID.trim()}`);
  };

  return (
    <form className="board-jump" onSubmit={onSubmit}>
      <label htmlFor="board-id">Board ID</label>
      <div className="board-jump-row">
        <input
          id="board-id"
          name="board-id"
          value={boardID}
          onChange={(event) => setBoardID(event.target.value)}
          placeholder="10001111111111111"
        />
        <button type="submit">Open Board</button>
      </div>
    </form>
  );
}

