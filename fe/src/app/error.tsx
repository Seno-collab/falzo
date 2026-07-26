"use client";

import { useEffect } from "react";
import { ErrorScreen } from "@/components/error-screen";

export default function ErrorPage({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    console.error("Unhandled frontend error", error);
  }, [error]);

  return (
    <ErrorScreen
      description="Something unexpected interrupted the game. Your room has not been intentionally closed."
      eyebrow="SOMETHING WENT WRONG"
      onRetry={reset}
      statusCode="500"
      title="The game hit a bad draw."
    />
  );
}
