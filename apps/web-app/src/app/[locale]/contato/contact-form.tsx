"use client";

import { useTranslations } from "next-intl";
import { useState, type FormEvent } from "react";

import styles from "./page.module.css";

type Status = "idle" | "sending" | "success" | "error";

export function ContactForm() {
  const t = useTranslations("contact");
  const [status, setStatus] = useState<Status>("idle");

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = event.currentTarget;
    const data = new FormData(form);

    setStatus("sending");

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
      setStatus("error");
    } catch {
      setStatus("error");
    }
  }

  if (status === "success") {
    return (
      <div className={styles.success} role="status">
        <p className={styles.successTitle}>{t("successTitle")}</p>
        <p className={styles.successText}>{t("successText")}</p>
        <button
          type="button"
          className={styles.successReset}
          onClick={() => setStatus("idle")}
        >
          {t("sendAnother")}
        </button>
      </div>
    );
  }

  return (
    <form className={styles.form} onSubmit={handleSubmit}>
      <div className={styles.field}>
        <label htmlFor="contact-name">{t("name")}</label>
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
        <label htmlFor="contact-email">{t("email")}</label>
        <input
          id="contact-email"
          name="email"
          type="email"
          autoComplete="email"
          required
        />
      </div>
      <div className={styles.field}>
        <label htmlFor="contact-message">{t("message")}</label>
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
          {status === "sending" ? t("sending") : t("submit")}
        </button>
        <p className={styles.hint}>{t("hint")}</p>
      </div>

      <p className={styles.error} role="alert" aria-live="polite">
        {status === "error" ? t("error") : ""}
      </p>
    </form>
  );
}
