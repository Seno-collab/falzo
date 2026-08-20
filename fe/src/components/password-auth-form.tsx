"use client";

import { FormEvent, useState } from "react";
import { useRouter } from "next/navigation";
import { ApiError, login, register } from "@/lib/api";
import { saveSession } from "@/lib/auth";

type AuthMode = "login" | "register";

const usernamePattern = /^[A-Za-z0-9_-]+$/;

export function PasswordAuthForm() {
  const router = useRouter();
  const [mode, setMode] = useState<AuthMode>("login");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [error, setError] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  function switchMode(nextMode: AuthMode) {
    setMode(nextMode);
    setError("");
    setPassword("");
    setConfirmPassword("");
    setShowPassword(false);
    setShowConfirmPassword(false);
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const normalizedUsername = username.trim().toLowerCase();

    if (normalizedUsername.length < 3 || normalizedUsername.length > 30 || !usernamePattern.test(normalizedUsername)) {
      setError("Username must be 3-30 characters using letters, numbers, _ or -.");
      return;
    }
    if (password.length < 8 || password.length > 72) {
      setError("Password must be between 8 and 72 characters.");
      return;
    }
    if (mode === "register" && password !== confirmPassword) {
      setError("Passwords do not match.");
      return;
    }

    setError("");
    setIsSubmitting(true);
    try {
      const result = mode === "register"
        ? await register(normalizedUsername, password)
        : await login(normalizedUsername, password);
      saveSession(result.username, result);
      router.replace("/dashboard");
    } catch (requestError) {
      if (requestError instanceof ApiError && requestError.code === "USERNAME_ALREADY_EXISTS") {
        setError("That username is already taken.");
      } else if (requestError instanceof ApiError && requestError.code === "INVALID_CREDENTIALS") {
        setError("Incorrect username or password.");
      } else if (requestError instanceof ApiError) {
        setError(requestError.message);
      } else {
        setError("Unable to continue right now. Please try again.");
      }
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <div className="password-auth">
      <div className="auth-mode-switch" aria-label="Account action">
        <button
          className={mode === "login" ? "is-active" : ""}
          type="button"
          aria-pressed={mode === "login"}
          onClick={() => switchMode("login")}
        >
          Sign in
        </button>
        <button
          className={mode === "register" ? "is-active" : ""}
          type="button"
          aria-pressed={mode === "register"}
          onClick={() => switchMode("register")}
        >
          Create account
        </button>
      </div>

      <form className="password-auth-form" onSubmit={submit}>
        <label className="auth-field">
          <span>Username</span>
          <input
            autoCapitalize="none"
            autoComplete="username"
            maxLength={30}
            minLength={3}
            name="username"
            pattern="[A-Za-z0-9_-]+"
            placeholder="your-player-name"
            required
            spellCheck={false}
            value={username}
            onChange={(event) => setUsername(event.target.value)}
          />
        </label>

        <label className="auth-field">
          <span>Password</span>
          <div className="auth-input-wrap">
            <input
              autoComplete={mode === "register" ? "new-password" : "current-password"}
              maxLength={72}
              minLength={8}
              name="password"
              placeholder="At least 8 characters"
              required
              type={showPassword ? "text" : "password"}
              value={password}
              onChange={(event) => setPassword(event.target.value)}
            />
            <button
              className="password-visibility"
              type="button"
              aria-label={showPassword ? "Hide password" : "Show password"}
              aria-pressed={showPassword}
              onClick={() => setShowPassword((visible) => !visible)}
            >
              <EyeIcon crossed={showPassword} />
            </button>
          </div>
        </label>

        {mode === "register" ? (
          <label className="auth-field">
            <span>Confirm password</span>
            <div className="auth-input-wrap">
              <input
                autoComplete="new-password"
                maxLength={72}
                minLength={8}
                name="confirm-password"
                placeholder="Repeat your password"
                required
                type={showConfirmPassword ? "text" : "password"}
                value={confirmPassword}
                onChange={(event) => setConfirmPassword(event.target.value)}
              />
              <button
                className="password-visibility"
                type="button"
                aria-label={showConfirmPassword ? "Hide password confirmation" : "Show password confirmation"}
                aria-pressed={showConfirmPassword}
                onClick={() => setShowConfirmPassword((visible) => !visible)}
              >
                <EyeIcon crossed={showConfirmPassword} />
              </button>
            </div>
          </label>
        ) : null}

        {mode === "register" ? (
          <p className="password-hint">Your username will be saved in lowercase.</p>
        ) : null}
        {error ? <p className="form-error" role="alert">{error}</p> : null}

        <button className="password-auth-submit" disabled={isSubmitting} type="submit">
          {isSubmitting
            ? "Please wait..."
            : mode === "register" ? "Create account" : "Sign in"}
        </button>
      </form>
    </div>
  );
}

function EyeIcon({ crossed }: { crossed: boolean }) {
  return (
    <svg aria-hidden="true" viewBox="0 0 24 24">
      <path d="M2.5 12s3.5-6 9.5-6 9.5 6 9.5 6-3.5 6-9.5 6-9.5-6-9.5-6Z" />
      <circle cx="12" cy="12" r="2.5" />
      {crossed ? <path d="m4 4 16 16" /> : null}
    </svg>
  );
}
