"use client";

import { Search, X } from "lucide-react";
import { Input } from "@/components/common/input";
import { Button } from "@/components/common/button";

export function SearchBar({
  value,
  onChange,
  placeholder,
}: {
  value: string;
  onChange: (nextValue: string) => void;
  placeholder: string;
}) {
  return (
    <div className="surface flex items-center gap-2 p-2">
      <span className="inline-flex h-9 w-9 items-center justify-center rounded-lg bg-[#edf4ff] text-[#3f6ea8]">
        <Search className="size-4" />
      </span>
      <Input
        className="h-9 border-none bg-transparent p-0 text-sm shadow-none focus:ring-0"
        onChange={(event) => onChange(event.target.value)}
        placeholder={placeholder}
        value={value}
      />
      {value ? (
        <Button
          aria-label="Clear search"
          onClick={() => onChange("")}
          size="icon"
          type="button"
          variant="ghost"
        >
          <X className="size-4" />
        </Button>
      ) : null}
    </div>
  );
}
