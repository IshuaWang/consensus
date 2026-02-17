import Link from "next/link";
import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { ApiRequestError, loginByEmail } from "@/lib/api";
import { ANSWER_TOKEN_COOKIE, ANSWER_TOKEN_MAX_AGE_SECONDS } from "@/lib/auth";
import { LoginForm, type LoginState } from "@/components/login-form";

type Props = {
  searchParams: Promise<{ from?: string }>;
};

function normalizeFrom(raw?: string): string {
  if (!raw) {
    return "/";
  }
  if (!raw.startsWith("/") || raw.startsWith("//") || raw.startsWith("/login")) {
    return "/";
  }
  return raw;
}

export default async function LoginPage({ searchParams }: Props) {
  const from = normalizeFrom((await searchParams).from);
  const authToken = (await cookies()).get(ANSWER_TOKEN_COOKIE)?.value;
  if (authToken) {
    redirect(from);
  }

  const loginAction = async (_state: LoginState, formData: FormData): Promise<LoginState> => {
    "use server";
    const email = String(formData.get("e_mail") ?? "").trim();
    const pass = String(formData.get("pass") ?? "");
    if (!email || !pass) {
      return { error: "Email and password are required." };
    }
    let accessToken = "";
    try {
      const resp = await loginByEmail({ e_mail: email, pass });
      if (!resp.access_token) {
        return { error: "Sign in failed: token missing from backend response." };
      }
      accessToken = resp.access_token;
      (await cookies()).set(ANSWER_TOKEN_COOKIE, accessToken, {
        httpOnly: true,
        sameSite: "lax",
        secure: process.env.NODE_ENV === "production",
        path: "/",
        maxAge: ANSWER_TOKEN_MAX_AGE_SECONDS
      });
    } catch (err) {
      if (err instanceof ApiRequestError) {
        return { error: `Sign in failed: ${err.message}` };
      }
      return { error: "Sign in failed: unknown error." };
    }
    redirect(from);
  };

  return (
    <main className="shell">
      <section className="page-head">
        <p className="hero-kicker">Authentication</p>
        <h1>Sign In For Forum Actions</h1>
        <p>
          After sign-in, `web-next` can create topics and replies through Answer forum APIs. Return
          path: <code>{from}</code>
        </p>
      </section>
      <section className="auth-wrap">
        <LoginForm action={loginAction} />
        <article className="panel auth-hint">
          <h2>Need password reset?</h2>
          <p>
            CLI reset:
            <code> go run ./cmd/answer passwd -C ./data -e admin@example.com</code>
          </p>
          <p>
            Original Answer UI is still available at{" "}
            <Link href="http://localhost:9080/users/login">http://localhost:9080/users/login</Link>.
          </p>
        </article>
      </section>
    </main>
  );
}
