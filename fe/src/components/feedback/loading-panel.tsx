import { LoaderCircle } from "lucide-react";
import { Skeleton } from "@/components/ui/skeleton";

export function LoadingPanel({
  title,
  description,
}: {
  title: string;
  description: string;
}) {
  return (
    <div className="app-panel-soft space-y-4 p-6">
      <div className="inline-flex items-center gap-2 text-sm font-medium text-[#416389]">
        <LoaderCircle className="size-4 animate-spin" />
        {title}
      </div>
      <p className="text-sm text-[#5b789b]">{description}</p>
      <div className="space-y-2.5">
        <Skeleton className="h-4 w-2/3" />
        <Skeleton className="h-4 w-4/5" />
        <Skeleton className="h-4 w-1/2" />
      </div>
    </div>
  );
}
