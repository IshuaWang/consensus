"use client";

import { useActionState } from "react";

export type LoginState = {
  error: string | null;
};

const initialState: LoginState = {
  error: null
};

type Props = {
  action: (state: LoginState, formData: FormData) => Promise<LoginState>;
};

export function LoginForm({ action }: Props) {
  const [state, formAction, pending] = useActionState(action, initialState);

  return (
    <form className="panel compose-form auth-form" action={formAction}>
      <h2>Sign In</h2>
      <p className="form-tip">
        This login is for forum write APIs in `web-next` and uses your Answer account.
      </p>
      <label htmlFor="login-email">Email</label>
      <input
        id="login-email"
        name="e_mail"
        type="email"
        autoComplete="email"
        required
        maxLength={500}
        placeholder="admin@example.com"
      />
      <label htmlFor="login-pass">Password</label>
      <input
        id="login-pass"
        name="pass"
        type="password"
        autoComplete="current-password"
        required
        minLength={8}
        maxLength={64}
      />
      {state.error && <p className="form-error">{state.error}</p>}
      <button disabled={pending} type="submit">
        {pending ? "Signing in..." : "Sign In"}
      </button>
    </form>
  );
}

