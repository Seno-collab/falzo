import { ErrorScreen } from "@/components/error-screen";

export default function NotFound() {
  return (
    <ErrorScreen
      description="The page may have moved, the room may have closed, or the link is incorrect."
      eyebrow="PAGE NOT FOUND"
      primaryLabel="Back home"
      statusCode="404"
      title="This card is not in the deck."
    />
  );
}
