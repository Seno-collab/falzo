"use client";

import { Banknote, CalendarDays, MapPinned, Palette, RotateCcw, Search } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

export type ItineraryFilterValues = {
  province: string;
  durationDays: string;
  budgetMax: string;
  travelStyle: string;
};

const provinceOptions = ["Phú Yên", "Đà Nẵng", "Yên Bái", "Ninh Bình"];
const styleOptions = ["biển", "chill", "chụp ảnh", "gia đình", "couple", "solo"];

export function ItineraryFilter({
  onChange,
  onReset,
  values,
}: Readonly<{
  values: ItineraryFilterValues;
  onChange: (values: ItineraryFilterValues) => void;
  onReset: () => void;
}>) {
  const update = (key: keyof ItineraryFilterValues, value: string) => {
    onChange({ ...values, [key]: value });
  };

  return (
    <section className="rounded-lg border border-[#d8e5e0] bg-white/94 p-4 shadow-[0_16px_42px_-36px_rgb(22_44_40/0.62)] backdrop-blur">
      <div className="mb-4 flex flex-col gap-3 border-b border-[#e8f0ed] pb-4 lg:flex-row lg:items-center lg:justify-between">
        <div>
          <p className="text-xs font-semibold uppercase tracking-[0.14em] text-[#397168]">
            Tìm lịch trình hợp gu
          </p>
          <p className="mt-1 text-sm text-[#667b76]">
            Chọn nhanh như feed social hoặc nhập bộ lọc chi tiết bên dưới.
          </p>
        </div>
        <div className="flex flex-wrap gap-2">
          {styleOptions.slice(0, 4).map((style) => {
            const isActive = values.travelStyle === style;

            return (
              <button
                className={
                  isActive
                    ? "rounded-full bg-[#17332e] px-3 py-1.5 text-xs font-semibold text-white shadow-sm"
                    : "rounded-full border border-[#d8e5e0] bg-[#f7fbf9] px-3 py-1.5 text-xs font-semibold text-[#315d55] transition hover:border-[#9fc7ba]"
                }
                key={style}
                onClick={() => update("travelStyle", isActive ? "" : style)}
                type="button"
              >
                #{style}
              </button>
            );
          })}
        </div>
      </div>

      <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-[1fr_0.8fr_0.8fr_0.9fr_auto]">
        <label className="space-y-1.5">
          <span className="inline-flex items-center gap-1.5 text-xs font-semibold uppercase tracking-[0.12em] text-[#6d7c78]">
            <MapPinned className="size-3.5" />
            Tỉnh/thành
          </span>
          <Input
            list="itinerary-provinces"
            onChange={(event) => update("province", event.target.value)}
            placeholder="Phú Yên"
            value={values.province}
          />
          <datalist id="itinerary-provinces">
            {provinceOptions.map((province) => (
              <option key={province} value={province} />
            ))}
          </datalist>
        </label>

        <label className="space-y-1.5">
          <span className="inline-flex items-center gap-1.5 text-xs font-semibold uppercase tracking-[0.12em] text-[#6d7c78]">
            <CalendarDays className="size-3.5" />
            Số ngày
          </span>
          <select
            className="h-11 w-full rounded-lg border border-input bg-white/90 px-3.5 text-sm font-semibold text-[#17332e] outline-none transition hover:border-[#b5cae2] focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/40"
            onChange={(event) => update("durationDays", event.target.value)}
            value={values.durationDays}
          >
            <option value="">Tất cả</option>
            {[1, 2, 3, 4, 5].map((day) => (
              <option key={day} value={day}>
                {day} ngày
              </option>
            ))}
          </select>
        </label>

        <label className="space-y-1.5">
          <span className="inline-flex items-center gap-1.5 text-xs font-semibold uppercase tracking-[0.12em] text-[#6d7c78]">
            <Banknote className="size-3.5" />
            Ngân sách tối đa
          </span>
          <Input
            inputMode="numeric"
            onChange={(event) => update("budgetMax", event.target.value)}
            placeholder="1500000"
            value={values.budgetMax}
          />
        </label>

        <label className="space-y-1.5">
          <span className="inline-flex items-center gap-1.5 text-xs font-semibold uppercase tracking-[0.12em] text-[#6d7c78]">
            <Palette className="size-3.5" />
            Style
          </span>
          <Input
            list="itinerary-styles"
            onChange={(event) => update("travelStyle", event.target.value)}
            placeholder="biển, chill"
            value={values.travelStyle}
          />
          <datalist id="itinerary-styles">
            {styleOptions.map((style) => (
              <option key={style} value={style} />
            ))}
          </datalist>
        </label>

        <div className="flex items-end gap-2 md:col-span-2 xl:col-span-1">
          <Button className="flex-1 xl:flex-none" type="submit" variant="default">
            <Search className="size-4" />
            Lọc
          </Button>
          <Button onClick={onReset} type="button" variant="outline">
            <RotateCcw className="size-4" />
          </Button>
        </div>
      </div>
    </section>
  );
}
