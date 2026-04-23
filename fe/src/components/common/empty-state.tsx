import { Compass } from "lucide-react";

export function EmptyState({ title, description }: { title: string; description: string }) {
  return (
    <div className="surface flex min-h-56 flex-col items-center justify-center gap-3 p-6 text-center">
      <span className="inline-flex h-12 w-12 items-center justify-center rounded-2xl bg-[#eaf2ff] text-[#2f5e95]">
        <Compass className="size-5" />
      </span>
      <h3 className="text-base font-semibold text-[#133050]">{title}</h3>
      <p className="max-w-md text-sm text-[#5e738f]">{description}</p>
    </div>
  );
}
