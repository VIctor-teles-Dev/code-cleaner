import type { Metadata } from "next";
import Link from "next/link";

import { listAllPosts } from "@/lib/admin";
import { requireSession } from "@/lib/auth";

import styles from "./admin.module.css";
import { LogoutButton } from "./logout-button";
import { PostList } from "./post-list";

export const metadata: Metadata = {
  title: "Admin — posts",
  robots: { index: false, follow: false },
};

export const dynamic = "force-dynamic";

export default async function AdminPage() {
  await requireSession();
  const posts = await listAllPosts();

  return (
    <main className={styles.page}>
      <div className={styles.container}>
        <div className={styles.header}>
          <div>
            <p className={styles.eyebrow}>$ ls blog/ --all</p>
            <h1 className={styles.title}>Posts</h1>
          </div>
          <div className={styles.headerActions}>
            <Link href="/admin/posts/new" className={styles.primary}>
              + novo post
            </Link>
            <LogoutButton />
          </div>
        </div>

        <PostList posts={posts} />
      </div>
    </main>
  );
}
