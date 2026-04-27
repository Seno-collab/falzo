import {
  getOptimizedScenicImage,
  getScenicImageUrl,
} from "@/lib/scenic-images";

type ScenicImageProps = {
  id: string;
  alt: string;
  className?: string;
  sizes?: string;
  loading?: "eager" | "lazy";
  fetchPriority?: "auto" | "high" | "low";
};

export function ScenicImage({
  id,
  alt,
  className,
  sizes = "(max-width: 768px) 100vw, (max-width: 1280px) 50vw, 33vw",
  loading = "lazy",
  fetchPriority = "auto",
}: ScenicImageProps) {
  const optimized = getOptimizedScenicImage(id);
  const remoteFallback = getScenicImageUrl(id);

  if (!optimized && !remoteFallback) {
    return null;
  }

  if (!optimized) {
    return (
      <img
        alt={alt}
        className={className}
        decoding="async"
        fetchPriority={fetchPriority}
        loading={loading}
        onError={(event) => {
          event.currentTarget.style.display = "none";
        }}
        src={remoteFallback ?? undefined}
      />
    );
  }

  return (
    <picture>
      <source sizes={sizes} srcSet={optimized.avifSrcSet} type="image/avif" />
      <source sizes={sizes} srcSet={optimized.webpSrcSet} type="image/webp" />
      <img
        alt={alt}
        className={className}
        decoding="async"
        fetchPriority={fetchPriority}
        loading={loading}
        onError={(event) => {
          if (
            remoteFallback &&
            !event.currentTarget.src.includes(remoteFallback)
          ) {
            event.currentTarget.src = remoteFallback;
            return;
          }

          event.currentTarget.style.display = "none";
        }}
        sizes={sizes}
        src={optimized.fallbackSrc ?? remoteFallback ?? undefined}
        srcSet={optimized.jpgSrcSet}
      />
    </picture>
  );
}
