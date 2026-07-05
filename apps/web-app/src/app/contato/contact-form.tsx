"use client";

import { useState, type FormEvent } from "react";

import styles from "./page.module.css";

type Status = "idle" | "sending" | "success" | "error";

export function ContactForm() {
  const [status, setStatus] = useState<Status>("idle");
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = event.currentTarget;
    const data = new FormData(form);

    setStatus("sending");
    setErrorMessage(null);

    try {
      const response = await fetch("/api/contact", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          name: data.get("name"),
          email: data.get("email"),
          message: data.get("message"),
        }),
      });

      if (response.ok) {
        setStatus("success");
        form.reset();
        return;
      }

      const body = await response.json().catch(() => null);
      setErrorMessage(
        body && typeof body.error === "string" ? body.error : null,
      );
      setStatus("error");
    } catch {
      setStatus("error");
    }
  }

  if (status === "success") {
    return (
      <div className={styles.success} role="status">
        <p className={styles.successTitle}>Mensagem enviada ✓</p>
        <p className={styles.successText}>
          Obrigado pelo contato! Respondo no email que você informou.
        </p>
        <button
          type="button"
          className={styles.successReset}
          onClick={() => setStatus("idle")}
        >
          Enviar outra mensagem
        </button>
      </div>
    );
  }

  return (
    <form className={styles.form} onSubmit={handleSubmit}>
      <div className={styles.field}>
        <label htmlFor="contact-name">Nome</label>
        <input
          id="contact-name"
          name="name"
          type="text"
          autoComplete="name"
          maxLength={200}
          required
        />
      </div>
      <div className={styles.field}>
        <label htmlFor="contact-email">Email</label>
        <input
          id="contact-email"
          name="email"
          type="email"
          autoComplete="email"
          required
        />
      </div>
      <div className={styles.field}>
        <label htmlFor="contact-message">Mensagem</label>
        <textarea
          id="contact-message"
          name="message"
          rows={6}
          maxLength={5000}
          required
        />
      </div>

      <div className={styles.actions}>
        <button
          type="submit"
          className={styles.submit}
          disabled={status === "sending"}
        >
          {status === "sending" ? "Enviando…" : "Enviar mensagem"}
        </button>
        <p className={styles.hint}>
          Sua mensagem fica registrada e eu recebo uma notificação.
        </p>
      </div>

      <p className={styles.error} role="alert" aria-live="polite">
        {status === "error"
          ? (errorMessage ??
            "Não foi possível enviar agora. Tente de novo em instantes.")
          : ""}
      </p>
    </form>
  );
}
