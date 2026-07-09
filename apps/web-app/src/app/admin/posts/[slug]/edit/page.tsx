import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";

import { getAnyPost } from "@/lib/admin";
import { requireSession } from "@/lib/auth";

import styles from "../../../admin.module.css";
import { PostEditor } from "../../../post-editor";

export const metadata: Metadata = {
  title: "Admin — editar post",
  robots: { index: false, follow: false },
};

export const dynamic = "force-dynamic";

interface EditPageProps {
  params: Promise<{ slug: string }>;
}

export default async function EditPostPage({ params }: EditPageProps) {
  await requireSession();
  const { slug } = await params;
  const post = await getAnyPost(slug);
  if (!post) {
    notFound();
  }

  return (
    <main className={styles.page}>
      <div className={styles.narrow}>
        <Link href="/admin" className={styles.back}>
          ← Voltar
        </Link>
        <h1 className={styles.title}>Editar post</h1>
        <PostEditor initial={post} />
      </div>
    </main>
  );
}
