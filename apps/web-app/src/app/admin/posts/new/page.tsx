import type { Metadata } from "next";
import Link from "next/link";

import { requireSession } from "@/lib/auth";

import styles from "../../admin.module.css";
import { PostEditor } from "../../post-editor";

export const metadata: Metadata = {
  title: "Admin — novo post",
  robots: { index: false, follow: false },
};

export const dynamic = "force-dynamic";

export default async function NewPostPage() {
  await requireSession();

  return (
    <main className={styles.page}>
      <div className={styles.narrow}>
        <Link href="/admin" className={styles.back}>
          ← Voltar
        </Link>
        <h1 className={styles.title}>Novo post</h1>
        <PostEditor />
      </div>
    </main>
  );
}
