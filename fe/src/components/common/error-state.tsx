import { TriangleAlert } from "lucide-react";
import { Button } from "@/components/common/button";

export function ErrorState({
  message,
  onRetry,
}: {
  message: string;
  onRetry?: () => void;
}) {
  return (
    <div className="surface flex min-h-56 flex-col items-center justify-center gap-3 p-6 text-center">
      <span className="inline-flex h-12 w-12 items-center justify-center rounded-2xl bg-[#fdecef] text-[#c83d4b]">
        <TriangleAlert className="size-5" />
      </span>
      <h3 className="text-base font-semibold text-[#133050]">Could not load data</h3>
      <p className="max-w-md text-sm text-[#5e738f]">{message}</p>
      {onRetry ? (
        <Button onClick={onRetry} type="button" variant="secondary">
          Retry
        </Button>
      ) : null}
    </div>
  );
}
