import { scenicImageManifest } from "@/generated/scenic-image-manifest";

const SCENIC_IMAGE_BY_ID: Record<string, string> = {
  "mu-cang-chai-dawn": "https://picsum.photos/seed/falzo-mu-cang-chai/1500/900",
  "ly-son-coast": "https://picsum.photos/seed/falzo-ly-son/900/1400",
  "kyoto-lantern-night":
    "https://picsum.photos/seed/falzo-kyoto-night/1600/1000",
  "swiss-lake-view": "https://picsum.photos/seed/falzo-swiss-lake/1000/1500",
  "istanbul-skyline": "https://picsum.photos/seed/falzo-istanbul/1400/900",
  "patagonia-trail": "https://picsum.photos/seed/falzo-patagonia/1800/1000",
  "mu-cang-chai": "https://picsum.photos/seed/falzo-mu-cang-chai/1500/900",
  "ly-son": "https://picsum.photos/seed/falzo-ly-son/900/1400",
  kyoto: "https://picsum.photos/seed/falzo-kyoto-night/1600/1000",
  interlaken: "https://picsum.photos/seed/falzo-swiss-lake/1000/1500",
  istanbul: "https://picsum.photos/seed/falzo-istanbul/1400/900",
  patagonia: "https://picsum.photos/seed/falzo-patagonia/1800/1000",
};

function toSrcSet(variants: Array<{ src: string; width: number }>) {
  return variants
    .map((variant) => `${variant.src} ${variant.width}w`)
    .join(", ");
}

export function getScenicImageUrl(id: string) {
  return SCENIC_IMAGE_BY_ID[id] ?? null;
}

export function getOptimizedScenicImage(id: string) {
  const entry = scenicImageManifest.images[id];
  if (!entry) {
    return null;
  }

  const fallbackVariant = entry.jpg[entry.jpg.length - 1];

  return {
    avifSrcSet: toSrcSet(entry.avif),
    webpSrcSet: toSrcSet(entry.webp),
    jpgSrcSet: toSrcSet(entry.jpg),
    fallbackSrc: fallbackVariant?.src ?? null,
    aspectRatio: entry.aspectRatio,
  };
}
