import type { Metadata } from "next";

import { ContactForm } from "./contact-form";
import styles from "./page.module.css";

export const metadata: Metadata = {
  title: "Contato",
  description:
    "Fale comigo: mande sua mensagem e eu respondo por email.",
};

export default function Contato() {
  return (
    <main className={styles.page}>
      <div className={styles.container}>
        <p className={styles.eyebrow}>$ mail victor</p>
        <h1 className={styles.title}>Contato</h1>
        <p className={styles.subtitle}>
          Tem um projeto em mente, uma vaga ou só quer trocar uma ideia sobre
          código? Me manda uma mensagem — eu leio todas e respondo por email.
        </p>
        <ContactForm />
      </div>
    </main>
  );
}
