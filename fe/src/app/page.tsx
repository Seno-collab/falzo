"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { getSession } from "@/lib/auth";
import styles from "./page.module.css";

const gameSteps = [
  {
    number: "01",
    title: "Gather the group",
    description: "Bring 4–12 friends together around one table.",
  },
  {
    number: "02",
    title: "Read your word",
    description: "Most players match. The Undercover gets a similar word.",
  },
  {
    number: "03",
    title: "Give a careful clue",
    description: "Sound convincing, watch reactions, and vote out the outsider.",
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

  return (
    <main className={styles.page}>
      <header className={styles.header}>
        <Link className={styles.brand} href="/">
          <span aria-hidden="true">F</span>
          falzo
        </Link>

        <nav className={styles.navigation} aria-label="Main navigation">
          <a href="#undercover">The game</a>
          <a href="#how-to-play">How it works</a>
        </nav>

        <div className={styles.headerAction}>
          {sessionReady ? (
            username ? (
              <Link className={styles.accountLink} href="/dashboard">
                <span className={styles.avatar} aria-hidden="true">{initial}</span>
                <span>{username}</span>
                <span aria-hidden="true">→</span>
              </Link>
            ) : (
              <Link className={styles.signInLink} href="/login">
                Sign in <span aria-hidden="true">↗</span>
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
            <span aria-hidden="true" /> SOCIAL DEDUCTION FOR REAL FRIENDS
          </p>
          <h1>One word is wrong.<br />One friend is lying.</h1>
          <p className={styles.heroDescription}>
            Undercover turns any table into a guessing game. Read your secret word, give a
            careful clue, and find the player who does not belong.
          </p>

          <div className={styles.heroActions}>
            <Link className={styles.primaryAction} href={username ? "/dashboard" : "/login"}>
              {username ? "Open game lobby" : "Enter the lobby"}
              <span aria-hidden="true">→</span>
            </Link>
            <a className={styles.textAction} href="#how-to-play">
              See how it works <span aria-hidden="true">↓</span>
            </a>
          </div>

          <dl className={styles.heroFacts} aria-label="Undercover game facts">
            <div><dt>4–12</dt><dd>players</dd></div>
            <div><dt>10–20</dt><dd>minutes</dd></div>
            <div><dt>Browser</dt><dd>no download</dd></div>
          </dl>
        </div>

        <div className={styles.gameBoard} aria-label="Preview of Undercover secret word cards">
          <div className={styles.boardTopline}>
            <span>UNDERCOVER</span>
            <span className={styles.developmentStatus}><i /> IN DEVELOPMENT</span>
          </div>

          <div className={styles.cardStage} aria-hidden="true">
            <div className={`${styles.wordCard} ${styles.backCard}`}>
              <span>FALZO</span>
              <strong>?</strong>
              <small>KEEP IT SECRET</small>
            </div>
            <div className={`${styles.wordCard} ${styles.civilianCard}`}>
              <span>YOUR WORD</span>
              <strong>FOREST</strong>
              <small>CIVILIAN</small>
            </div>
            <div className={`${styles.wordCard} ${styles.undercoverCard}`}>
              <span>YOUR WORD</span>
              <strong>JUNGLE</strong>
              <small>UNDERCOVER</small>
            </div>
          </div>

          <div className={styles.boardFooter}>
            <div className={styles.playerStack} aria-hidden="true">
              <span>L</span><span>M</span><span>A</span><span>+3</span>
            </div>
            <span>Who has the wrong word?</span>
          </div>
        </div>
      </section>

      <section className={styles.gameStrip} aria-label="Game highlights">
        <span>NO APP STORE</span>
        <i aria-hidden="true">✦</i>
        <span>ONE SECRET WORD</span>
        <i aria-hidden="true">✦</i>
        <span>ENDLESS ACCUSATIONS</span>
        <i aria-hidden="true">✦</i>
        <span>MADE FOR GAME NIGHT</span>
      </section>

      <section className={styles.gameIntro} id="undercover">
        <div className={styles.introNumber}>01</div>
        <div className={styles.introCopy}>
          <p className={styles.darkEyebrow}>MEET THE FIRST GAME</p>
          <h2>Simple rules.<br />Suspicious friends.</h2>
        </div>
        <div className={styles.introDescription}>
          <p>
            Undercover is easy to explain and hard to master. You do not need a board, a deck,
            or a long tutorial—just a group that thinks it knows each other.
          </p>
          <span>Social deduction · Word game · Party game</span>
        </div>
      </section>

      <section className={styles.howToPlay} id="how-to-play" aria-labelledby="how-title">
        <div className={styles.sectionTitle}>
          <p>HOW A ROUND WORKS</p>
          <h2 id="how-title">Three moves between<br />friends and suspicion.</h2>
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

        <div className={styles.bottomCta}>
          <div>
            <span>FALZO GAME 01</span>
            <h2>The table is almost ready.</h2>
          </div>
          <Link href={username ? "/dashboard" : "/login"}>
            {username ? "Go to your lobby" : "Join Falzo"} <span aria-hidden="true">→</span>
          </Link>
        </div>
      </section>

      <footer className={styles.footer}>
        <Link className={styles.brand} href="/">
          <span aria-hidden="true">F</span>
          falzo
        </Link>
        <p>Games for dinners, road trips, and nights with friends.</p>
        <small>© 2026 FALZO</small>
      </footer>
    </main>
  );
}
