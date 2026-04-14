export type ScenicImageVariant = {
  src: string;
  width: number;
};

export type ScenicImageManifestEntry = {
  aspectRatio: number;
  avif: ScenicImageVariant[];
  webp: ScenicImageVariant[];
  jpg: ScenicImageVariant[];
};

export type ScenicImageManifest = {
  generatedAt: string;
  widths: number[];
  images: Record<string, ScenicImageManifestEntry>;
};
