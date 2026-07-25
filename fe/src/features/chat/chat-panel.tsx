"use client";

import { type FormEvent, useId, useRef, useState, useEffect } from "react";
import styles from "./chat-panel.module.css";

export type ChatMessage = {
  id: string;
  sender: string;
  text: string;
  time: string;
  own?: boolean;
  system?: boolean;
};

type ChatPanelProps = {
  title: string;
  subtitle: string;
  currentUsername: string;
  initialMessages: readonly ChatMessage[];
  presence?: "online" | "offline" | "room";
  contextLabel?: string;
  inputPlaceholder?: string;
  className?: string;
  onClose?: () => void;
};

export function ChatPanel({
  title,
  subtitle,
  currentUsername,
  initialMessages,
  presence = "online",
  contextLabel,
  inputPlaceholder = "Write a message…",
  className,
  onClose,
}: ChatPanelProps) {
  const inputId = useId();
  const [messages, setMessages] = useState<ChatMessage[]>(() => [...initialMessages]);
  const [draftMessage, setDraftMessage] = useState("");
  const chatEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    chatEndRef.current?.scrollIntoView({ block: "nearest" });
  }, [messages]);

  function sendMessage(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const text = draftMessage.trim();
    if (!text) return;

    setMessages((current) => [
      ...current,
      {
        id: `local-${Date.now()}`,
        sender: currentUsername,
        text,
        time: new Intl.DateTimeFormat("en", {
          hour: "2-digit",
          minute: "2-digit",
        }).format(new Date()),
        own: true,
      },
    ]);
    setDraftMessage("");
  }

  return (
    <section className={`${styles.panel} ${className ?? ""}`} aria-label={`Chat with ${title}`}>
      <header className={styles.header}>
        <div className={styles.identity}>
          <span className={`${styles.presenceDot} ${styles[presence]}`} aria-hidden="true" />
          <div>
            <strong>{title}</strong>
            <small>{subtitle}</small>
          </div>
        </div>

        <div className={styles.headerActions}>
          {contextLabel && <span>{contextLabel}</span>}
          {onClose && (
            <button aria-label={`Close chat with ${title}`} onClick={onClose} type="button">
              ×
            </button>
          )}
        </div>
      </header>

      <div className={styles.messages} aria-live="polite">
        {messages.map((message) => (
          message.system ? (
            <div className={styles.systemMessage} key={message.id}>
              <span aria-hidden="true">i</span>
              <p>{message.text}</p>
            </div>
          ) : (
            <article
              className={`${styles.message} ${message.own ? styles.ownMessage : ""}`}
              key={message.id}
            >
              <div className={styles.messageMeta}>
                <strong>{message.own ? "You" : message.sender}</strong>
                <span>{message.time}</span>
              </div>
              <p>{message.text}</p>
            </article>
          )
        ))}
        <div ref={chatEndRef} />
      </div>

      <form className={styles.form} onSubmit={sendMessage}>
        <label className={styles.srOnly} htmlFor={inputId}>
          Message {title}
        </label>
        <input
          autoComplete="off"
          id={inputId}
          maxLength={240}
          onChange={(event) => setDraftMessage(event.target.value)}
          placeholder={inputPlaceholder}
          type="text"
          value={draftMessage}
        />
        <button disabled={!draftMessage.trim()} type="submit">
          <span aria-hidden="true">↑</span>
          <span className={styles.srOnly}>Send message</span>
        </button>
      </form>
    </section>
  );
}
