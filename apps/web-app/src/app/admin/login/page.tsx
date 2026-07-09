import type { Metadata } from "next";
import { redirect } from "next/navigation";

import { isAuthenticated } from "@/lib/auth";

import styles from "../admin.module.css";
import { LoginForm } from "./login-form";

export const metadata: Metadata = {
  title: "Admin — entrar",
  robots: { index: false, follow: false },
};

export const dynamic = "force-dynamic";

export default async function LoginPage() {
  // Já autenticado: vai direto para o painel.
  if (await isAuthenticated()) {
    redirect("/admin");
  }

  return (
    <main className={styles.page}>
      <div className={styles.narrow}>
        <p className={styles.eyebrow}>$ sudo login</p>
        <h1 className={styles.title}>Admin do blog</h1>
        <p className={styles.subtitle}>Área restrita. Informe a senha.</p>
        <LoginForm />
      </div>
    </main>
  );
}
