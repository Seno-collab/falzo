"use client";

import { useEffect } from "react";
import { ErrorScreen } from "@/components/error-screen";
import styles from "@/components/error-screen.module.css";

export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    console.error("Unhandled root frontend error", error);
  }, [error]);

  return (
    <html lang="en">
      <body className={styles.errorBody}>
        <ErrorScreen
          description="Falzo could not load the application shell. Try again or return home to start fresh."
          eyebrow="APPLICATION ERROR"
          onRetry={reset}
          statusCode="500"
          title="We could not set the table."
        />
      </body>
    </html>
  );
}
