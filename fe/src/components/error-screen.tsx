"use client";

import Link from "next/link";
import styles from "./error-screen.module.css";

type ErrorScreenProps = {
  statusCode: "403" | "404" | "500";
  eyebrow: string;
  title: string;
  description: string;
  onRetryAction?: () => void;
  primaryHref?: string;
  primaryLabel?: string;
};

export function ErrorScreen({
  statusCode,
  eyebrow,
  title,
  description,
  onRetryAction,
  primaryHref = "/",
  primaryLabel = "Back home",
}: ErrorScreenProps) {
  return (
    <main className={styles.page}>
      <div className={styles.orbit} aria-hidden="true" />

      <section className={styles.content}>
        <Link className={styles.brand} href="/">
          <span aria-hidden="true">F</span>
          falzo
        </Link>

        <div className={styles.errorMark} aria-hidden="true">
          <span>{statusCode}</span>
          <div>
            <small>FALZO</small>
            <strong>?</strong>
            <small>NO MATCH</small>
          </div>
        </div>

        <p className={styles.eyebrow}>{eyebrow}</p>
        <h1>{title}</h1>
        <p className={styles.description}>{description}</p>

        <div className={styles.actions}>
          {onRetryAction && (
            <button
              className={styles.primaryAction}
              onClick={onRetryAction}
              type="button"
            >
              Try again
            </button>
          )}
          <Link
            className={
              onRetryAction ? styles.secondaryAction : styles.primaryAction
            }
            href={primaryHref}
          >
            {primaryLabel}
          </Link>
          {primaryHref !== "/" && (
            <Link className={styles.textAction} href="/">
              Home
            </Link>
          )}
        </div>
      </section>
    </main>
  );
}
