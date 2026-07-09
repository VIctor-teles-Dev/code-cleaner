"use client";

import { useRouter } from "next/navigation";
import { useState, type FormEvent } from "react";

import styles from "../admin.module.css";

type Status = "idle" | "sending" | "error";

export function LoginForm() {
  const router = useRouter();
  const [password, setPassword] = useState("");
  const [status, setStatus] = useState<Status>("idle");
  const [error, setError] = useState<string | null>(null);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setStatus("sending");
    setError(null);
    try {
      const res = await fetch("/api/admin/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ password }),
      });
      if (res.ok) {
        router.push("/admin");
        router.refresh();
        return;
      }
      const body = await res.json().catch(() => null);
      setError(
        body && typeof body.error === "string" ? body.error : "Não autorizado.",
      );
      setStatus("error");
    } catch {
      setError("Não foi possível entrar agora.");
      setStatus("error");
    }
  }

  return (
    <form className={styles.form} onSubmit={handleSubmit}>
      <div className={styles.field}>
        <label htmlFor="admin-password">senha</label>
        <input
          id="admin-password"
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          autoComplete="current-password"
          autoFocus
          required
        />
      </div>
      <div className={styles.actions}>
        <button
          type="submit"
          className={styles.primary}
          disabled={status === "sending" || password === ""}
        >
          {status === "sending" ? "Entrando…" : "Entrar"}
        </button>
      </div>
      <p className={styles.error} role="alert" aria-live="polite">
        {status === "error" ? error : ""}
      </p>
    </form>
  );
}
