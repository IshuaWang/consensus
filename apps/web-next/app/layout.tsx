import type { Metadata } from "next";
import Link from "next/link";
import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { Manrope, Source_Sans_3 } from "next/font/google";
import { ANSWER_TOKEN_COOKIE } from "@/lib/auth";
import "./globals.css";

const headingFont = Manrope({
  subsets: ["latin"],
  variable: "--font-heading",
  weight: ["500", "700", "800"]
});

const bodyFont = Source_Sans_3({
  subsets: ["latin"],
  variable: "--font-body",
  weight: ["400", "500", "700"]
});

export const metadata: Metadata = {
  title: "Consensus Forum",
  description: "Forum-first knowledge synthesis workspace."
};

export default async function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  const hasToken = Boolean((await cookies()).get(ANSWER_TOKEN_COOKIE)?.value);
  const logoutAction = async () => {
    "use server";
    (await cookies()).delete(ANSWER_TOKEN_COOKIE);
    redirect("/");
  };
  return (
    <html lang="en">
      <body className={`${headingFont.variable} ${bodyFont.variable}`}>
        <header className="topbar">
          <div className="topbar-inner">
            <Link className="topbar-brand" href="/">
              Consensus
            </Link>
            <nav className="topbar-nav">
              <Link href="/categories/demo-category-id">Demo Category</Link>
              {hasToken ? (
                <form action={logoutAction}>
                  <button type="submit">Sign Out</button>
                </form>
              ) : (
                <Link href="/login">Sign In</Link>
              )}
            </nav>
          </div>
        </header>
        {children}
      </body>
    </html>
  );
}
