"use client";

import { type SubmitEvent, useEffect, useId, useRef, useState } from "react";
import styles from "./chat-panel.module.css";

export type ChatMessage = {
  id: string;
  sender: string;
  text: string;
  time: string;
  sentAt?: string;
  own?: boolean;
  system?: boolean;
};

type ChatPanelProps = {
  title: string;
  subtitle: string;
  messages?: readonly ChatMessage[];
  initialMessages?: readonly ChatMessage[];
  currentUsername?: string;
  connected?: boolean;
  onSendMessageAction?: (text: string) => boolean;
  presence?: "online" | "offline" | "room";
  contextLabel?: string;
  inputPlaceholder?: string;
  disabledPlaceholder?: string;
  className?: string;
  onCloseAction?: () => void;
};

export function ChatPanel({
  title,
  subtitle,
  messages,
  initialMessages = [],
  currentUsername = "You",
  connected = true,
  onSendMessageAction,
  presence = "online",
  contextLabel,
  inputPlaceholder = "Write a message…",
  disabledPlaceholder = "Reconnecting…",
  className,
  onCloseAction,
}: ChatPanelProps) {
  const inputId = useId();
  const [localMessages, setLocalMessages] = useState<ChatMessage[]>(() => [...initialMessages]);
  const [draftMessage, setDraftMessage] = useState("");
  const chatEndRef = useRef<HTMLDivElement>(null);
  const visibleMessages = messages ?? localMessages;

  useEffect(() => {
    chatEndRef.current?.scrollIntoView({ block: "nearest" });
  }, [visibleMessages]);

  function sendMessage(event: SubmitEvent<HTMLFormElement>) {
    event.preventDefault();
    const text = draftMessage.trim();
    if (!text) return;

    if (onSendMessageAction) {
      if (onSendMessageAction(text)) setDraftMessage("");
      return;
    }

    setLocalMessages((current) => [...current, {
      id: `local-${Date.now()}`,
      sender: currentUsername,
      text,
      time: new Intl.DateTimeFormat("en", { hour: "2-digit", minute: "2-digit" }).format(new Date()),
      own: true,
    }]);
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
          {onCloseAction && (
            <button aria-label={`Close chat with ${title}`} onClick={onCloseAction} type="button">
              ×
            </button>
          )}
        </div>
      </header>

      <div className={styles.messages} aria-live="polite">
        {visibleMessages.length === 0 && (
          <div className={styles.systemMessage}>
            <span aria-hidden="true">i</span>
            <p>No messages yet. Say hello to the room.</p>
          </div>
        )}
        {visibleMessages.map((message) => (
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
          disabled={!connected}
          id={inputId}
          maxLength={500}
          onChange={(event) => setDraftMessage(event.target.value)}
          placeholder={connected ? inputPlaceholder : disabledPlaceholder}
          type="text"
          value={draftMessage}
        />
        <button disabled={!connected || !draftMessage.trim()} type="submit">
          <span aria-hidden="true">↑</span>
          <span className={styles.srOnly}>Send message</span>
        </button>
      </form>
    </section>
  );
}
