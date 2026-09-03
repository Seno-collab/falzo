"use client";

import Script from "next/script";
import { useRouter } from "next/navigation";
import { useRef, useState } from "react";
import { googleLogin } from "@/lib/api";
import { saveSession } from "@/lib/auth";

type GoogleCredentialResponse = {
  credential?: string;
};

type GoogleIdentityServices = {
  accounts: {
    id: {
      initialize(options: {
        client_id: string;
        callback(response: GoogleCredentialResponse): void;
      }): void;
      renderButton(
        parent: HTMLElement,
        options: {
          type: "standard";
          theme: "outline";
          size: "large";
          text: "continue_with";
          shape: "pill";
          logo_alignment: "left";
          width: number;
        },
      ): void;
    };
  };
};

declare global {
  interface Window {
    google?: GoogleIdentityServices;
  }
}

const googleClientID = process.env.NEXT_PUBLIC_GOOGLE_CLIENT_ID ?? "";

export function GoogleLoginButton() {
  const router = useRouter();
  const buttonRef = useRef<HTMLDivElement>(null);
  const [error, setError] = useState("");
  const [isReady, setIsReady] = useState(false);
  const [loadFailed, setLoadFailed] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  function initializeGoogleButton() {
    if (!googleClientID) {
      setLoadFailed(true);
      setError("Google login is not configured.");
      return;
    }
    if (!window.google || !buttonRef.current) return;

    const buttonWidth = Math.min(
      400,
      Math.floor(buttonRef.current.clientWidth),
    );
    buttonRef.current.replaceChildren();
    window.google.accounts.id.initialize({
      client_id: googleClientID,
      callback: async ({ credential }) => {
        if (!credential) {
          setError("Google did not return a credential.");
          return;
        }

        setError("");
        setIsSubmitting(true);
        try {
          const result = await googleLogin(credential);
          saveSession(result.username, result);
          router.replace("/dashboard");
        } catch {
          setError("Unable to sign in with Google. Please try again.");
          setIsSubmitting(false);
        }
      },
    });
    window.google.accounts.id.renderButton(buttonRef.current, {
      type: "standard",
      theme: "outline",
      size: "large",
      text: "continue_with",
      shape: "pill",
      logo_alignment: "left",
      width: buttonWidth,
    });
    setLoadFailed(false);
    setIsReady(true);
  }

  return (
    <div className="google-login">
      <Script
        src="https://accounts.google.com/gsi/client"
        strategy="afterInteractive"
        onReady={initializeGoogleButton}
        onError={() => {
          setLoadFailed(true);
          setError("Unable to load Google sign-in.");
        }}
      />
      <div
        className={`google-button-shell${isReady ? " is-ready" : ""}${loadFailed ? " has-error" : ""}${isSubmitting ? " is-submitting" : ""}`}
      >
        <div className="google-button-host" ref={buttonRef} />
        {!isReady ? (
          <span className="google-button-skeleton">
            {loadFailed ? "Google sign-in unavailable" : "Loading Google…"}
          </span>
        ) : null}
        {isSubmitting ? (
          <span className="google-button-progress" role="status">
            <span className="google-button-spinner" aria-hidden="true" />
            Signing you in…
          </span>
        ) : null}
      </div>
      <p className="google-login-note">
        <span aria-hidden="true">◆</span> Secure sign-in. We never see your
        Google password.
      </p>
      {error ? (
        <div className="auth-notification" role="alert">
          <span className="auth-notification-icon" aria-hidden="true">
            !
          </span>
          <span className="auth-notification-content">
            <strong>Google sign-in failed</strong>
            <span>{error}</span>
          </span>
        </div>
      ) : null}
    </div>
  );
}
