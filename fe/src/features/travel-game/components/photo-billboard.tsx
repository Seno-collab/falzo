"use client";

import { Html } from "@react-three/drei";
import { getScenicImageUrl } from "@/lib/scenic-images";

type PhotoBillboardProps = {
  imageId: string;
  selected: boolean;
  title: string;
};

export function PhotoBillboard({
  imageId,
  selected,
  title,
}: Readonly<PhotoBillboardProps>) {
  const imageUrl = getScenicImageUrl(imageId);

  if (!imageUrl) {
    return null;
  }

  return (
    <Html
      center
      className="pointer-events-none select-none"
      distanceFactor={selected ? 4.2 : 5.8}
      position={[selected ? 0.64 : 0.36, selected ? 1.22 : 0.82, -0.22]}
      transform
    >
      <div
        className={[
          "overflow-hidden rounded-lg border bg-[#071b18]/82 shadow-[0_22px_48px_-24px_rgb(0_0_0/0.92)] backdrop-blur",
          selected
            ? "w-52 border-[#f7c948]/62"
            : "w-24 border-white/16 opacity-80",
        ].join(" ")}
      >
        <div className="relative">
          <img
            alt=""
            className={[
              "w-full object-cover",
              selected ? "aspect-[16/9]" : "aspect-[4/3]",
            ].join(" ")}
            decoding="async"
            loading="lazy"
            src={imageUrl}
          />
          <div className="absolute inset-x-0 bottom-0 bg-[linear-gradient(0deg,rgba(6,22,23,0.86),rgba(6,22,23,0))] px-2 pb-1 pt-5">
            <p
              className={[
                "truncate font-bold text-white",
                selected ? "text-[0.68rem]" : "text-[0.58rem]",
              ].join(" ")}
            >
              {title}
            </p>
          </div>
        </div>
      </div>
    </Html>
  );
}
