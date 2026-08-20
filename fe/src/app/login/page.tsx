import Link from "next/link";
import { GoogleLoginButton } from "@/components/google-login-button";
import { PasswordAuthForm } from "@/components/password-auth-form";

export default function LoginPage() {
  return (
    <main className="login-page">
      <Link className="site-brand login-brand" href="/">
        <span className="brand-mark" aria-hidden="true">F</span>
        <span>falzo</span>
      </Link>

      <section className="login-layout">
        <div className="login-story">
          <Link className="back-link" href="/">← Back home</Link>
          <p className="eyebrow"><span aria-hidden="true">●</span> THE LOBBY IS OPEN</p>
          <h1>Good games start with good company.</h1>
          <p>Sign in once, then bring Falzo to the next dinner, road trip, or quiet night that needs a little chaos.</p>

          <div className="login-mini-card" aria-hidden="true">
            <div className="mini-card-icon">?</div>
            <div><strong>Undercover</strong><span>Bluff. Guess. Reveal.</span></div>
            <span className="mini-card-players">4–12</span>
          </div>
        </div>

        <div className="auth-card">
          <span className="auth-icon" aria-hidden="true">✦</span>
          <p className="eyebrow">WELCOME TO FALZO</p>
          <h2>Enter the game</h2>
          <p className="muted">Create a player account, sign in with your password, or continue with Google.</p>
          <PasswordAuthForm />
          <div className="auth-divider"><span>or</span></div>
          <GoogleLoginButton />
          <p className="auth-terms">By continuing, you agree to play nice.</p>
        </div>
      </section>
    </main>
  );
}
