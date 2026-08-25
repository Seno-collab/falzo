"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { getSession } from "@/lib/auth";
import styles from "./page.module.css";

const gameSteps = [
  {
    number: "01",
    title: "Join a room",
    description: "Bring 4–12 friends together and let everyone take a seat.",
  },
  {
    number: "02",
    title: "Read your word",
    description:
      "Most players share one word. The Undercover gets a similar one.",
  },
  {
    number: "03",
    title: "Talk and vote",
    description:
      "Give a clue, listen carefully, then vote out the suspicious player.",
  },
] as const;

export default function HomePage() {
  const [username, setUsername] = useState<string | null>(null);
  const [sessionReady, setSessionReady] = useState(false);

  useEffect(() => {
    setUsername(getSession()?.username ?? null);
    setSessionReady(true);
  }, []);

  const initial = username?.trim().charAt(0).toUpperCase() || "P";
  const lobbyHref = username ? "/dashboard" : "/login";

  return (
    <main className={styles.page}>
      <header className={styles.header}>
        <Link className={styles.brand} href="/" aria-label="Falzo home">
          <span aria-hidden="true">F</span>
          falzo
        </Link>

        <div className={styles.headerAction}>
          {sessionReady ? (
            username ? (
              <Link className={styles.accountLink} href="/dashboard">
                <span className={styles.avatar} aria-hidden="true">
                  {initial}
                </span>
                <span className={styles.accountName}>{username}</span>
              </Link>
            ) : (
              <Link className={styles.signInLink} href="/login">
                Sign in
              </Link>
            )
          ) : (
            <span className={styles.sessionPlaceholder} aria-hidden="true" />
          )}
        </div>
      </header>

      <section className={styles.hero}>
        <div className={styles.heroCopy}>
          <p className={styles.eyebrow}>
            <span aria-hidden="true" /> PARTY GAME FOR FRIENDS
          </p>
          <h1>Find the friend with the wrong word.</h1>
          <p className={styles.description}>
            Undercover is a simple social deduction game. Give a careful clue,
            read the room, and work out who does not belong.
          </p>

          <div className={styles.heroActions}>
            <Link className={styles.primaryAction} href={lobbyHref}>
              {username ? "View rooms" : "Start playing"}
              <span aria-hidden="true">→</span>
            </Link>
            <a className={styles.secondaryAction} href="#how-to-play">
              How to play
            </a>
          </div>

          <div className={styles.gameFacts} aria-label="Game details">
            <span>
              <strong>4–12</strong> players
            </span>
            <span>
              <strong>10–20</strong> minutes
            </span>
            <span>
              <strong>Free</strong> in browser
            </span>
          </div>
        </div>

        <div
          className={styles.gamePreview}
          aria-label="Undercover game preview"
        >
          <div className={styles.previewHeader}>
            <div>
              <span className={styles.previewLabel}>GAME 01</span>
              <h2>Undercover</h2>
            </div>
            <span className={styles.status}>
              <i aria-hidden="true" /> IN DEVELOPMENT
            </span>
          </div>

          <div className={styles.secretWord}>
            <span>YOUR SECRET WORD</span>
            <strong>FOREST</strong>
            <small>Keep it hidden from the table</small>
          </div>

          <div className={styles.previewFooter}>
            <div className={styles.players} aria-hidden="true">
              <span>L</span>
              <span>M</span>
              <span>A</span>
              <span>+3</span>
            </div>
            <p>One player has a different word.</p>
          </div>
        </div>
      </section>

      <section
        className={styles.howToPlay}
        id="how-to-play"
        aria-labelledby="how-title"
      >
        <div className={styles.sectionHeading}>
          <p>HOW TO PLAY</p>
          <h2 id="how-title">
            Easy to learn.
            <br />
            Hard to fake.
          </h2>
        </div>

        <ol className={styles.steps}>
          {gameSteps.map((step) => (
            <li key={step.number}>
              <span>{step.number}</span>
              <div>
                <h3>{step.title}</h3>
                <p>{step.description}</p>
              </div>
            </li>
          ))}
        </ol>
      </section>

      <section className={styles.cta}>
        <div>
          <span>READY FOR GAME NIGHT?</span>
          <h2>Take a seat and test your friends.</h2>
        </div>
        <Link href={lobbyHref}>
          {username ? "Open lobby" : "Join Falzo"}
          <span aria-hidden="true">→</span>
        </Link>
      </section>

      <footer className={styles.footer}>
        <Link className={styles.brand} href="/">
          <span aria-hidden="true">F</span>
          falzo
        </Link>
        <p>Simple games for real friends.</p>
        <small>© 2026 FALZO</small>
      </footer>
    </main>
  );
}
