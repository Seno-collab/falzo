Drop your original scenic images into this folder and run:

  pnpm images:optimize

Naming recommendation:
- mu-cang-chai.jpg
- ly-son.jpg
- kyoto.jpg
- interlaken.jpg
- istanbul.jpg
- patagonia.jpg

The optimizer will generate AVIF/WebP/JPG variants into:
  public/images/optimized/scenic

And update runtime manifest:
  src/generated/scenic-image-manifest.ts
