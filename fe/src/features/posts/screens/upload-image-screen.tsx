"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  ArrowLeft,
  Camera,
  Check,
  Compass,
  ImagePlus,
  Loader2,
  Map,
  MapPin,
  RefreshCcw,
  Trash2,
  Upload,
} from "lucide-react";
import { useRouter } from "next/navigation";
import { useEffect, useMemo, useRef, useState } from "react";
import { toast } from "sonner";
import { AppTopbar } from "@/components/layout/app-topbar";
import { PageShell } from "@/components/layout/page-shell";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  clearAuthSession,
  getApiErrorMessage,
  getMeApi,
  hasAuthSession,
} from "@/features/auth/api";
import { createPostApi, uploadImageApi } from "@/features/posts/api";
import type { UploadedImage } from "@/features/posts/types";
import { ROUTES } from "@/lib/routes";

type FormState = {
  caption: string;
  locationName: string;
  latitude: string;
  longitude: string;
};

const initialForm: FormState = {
  caption: "",
  locationName: "",
  latitude: "10.7769",
  longitude: "106.7009",
};
const maxImageSize = 10 * 1024 * 1024;
const acceptedImageTypes = new Set([
  "image/jpeg",
  "image/png",
  "image/webp",
  "image/jpg",
  "image/svg+xml",
  "image/gif",
]);

function parseCoordinate(value: string) {
  const normalized = Number(value);
  return Number.isFinite(normalized) ? normalized : Number.NaN;
}

function isSameFile(left: File | null, right: File) {
  if (!left) {
    return false;
  }

  return (
    left.name === right.name &&
    left.size === right.size &&
    left.lastModified === right.lastModified
  );
}

function formatFileSize(size: number) {
  if (size < 1024 * 1024) {
    return `${Math.ceil(size / 1024)} KB`;
  }

  return `${(size / 1024 / 1024).toFixed(1)} MB`;
}

function getImageValidationError(file: File) {
  if (!acceptedImageTypes.has(file.type)) {
    return "Only JPG, PNG, WebP, SVG, or GIF images are supported.";
  }

  if (file.size > maxImageSize) {
    return `Image must be smaller than ${formatFileSize(maxImageSize)}.`;
  }

  return null;
}

export function UploadImageScreen() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const selectedFileRef = useRef<File | null>(null);
  const heroFileInputRef = useRef<HTMLInputElement | null>(null);
  const formFileInputRef = useRef<HTMLInputElement | null>(null);
  const [isSessionChecking, setIsSessionChecking] = useState(true);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [uploadedImage, setUploadedImage] = useState<UploadedImage | null>(
    null,
  );
  const [form, setForm] = useState<FormState>(initialForm);

  const previewUrl = useMemo(() => {
    if (!selectedFile) {
      return null;
    }

    return URL.createObjectURL(selectedFile);
  }, [selectedFile]);

  useEffect(() => {
    document.title = "Upload Image | Falzo";

    if (!hasAuthSession()) {
      router.replace(ROUTES.login);
      return;
    }

    let disposed = false;

    const validateSession = async () => {
      try {
        await getMeApi();
        if (!disposed) {
          setIsSessionChecking(false);
        }
      } catch {
        if (disposed) {
          return;
        }

        clearAuthSession();
        router.replace(ROUTES.login);
      }
    };

    void validateSession();

    return () => {
      disposed = true;
    };
  }, [router]);

  useEffect(() => {
    return () => {
      if (previewUrl) {
        URL.revokeObjectURL(previewUrl);
      }
    };
  }, [previewUrl]);

  const uploadMutation = useMutation({
    mutationFn: uploadImageApi,
    onError: (error, file) => {
      if (isSameFile(selectedFileRef.current, file)) {
        setUploadedImage(null);
      }
      toast.error(getApiErrorMessage(error));
    },
    onSuccess: (image, file) => {
      if (!isSameFile(selectedFileRef.current, file)) {
        return;
      }

      setUploadedImage(image);
      toast.success("Image uploaded.");
    },
  });

  const publishMutation = useMutation({
    mutationFn: async () => {
      if (!selectedFile && !uploadedImage?.url) {
        throw new Error("Choose an image before publishing.");
      }

      const image = uploadedImage ?? (await uploadSelectedFile());
      const latitude = parseCoordinate(form.latitude);
      const longitude = parseCoordinate(form.longitude);
      if (!Number.isFinite(latitude) || !Number.isFinite(longitude)) {
        throw new TypeError("Latitude and longitude must be valid numbers.");
      }

      return createPostApi({
        image_url: image.url,
        caption: form.caption,
        location_name: form.locationName,
        latitude,
        longitude,
      });
    },
    onError: (error) => {
      toast.error(getApiErrorMessage(error));
    },
    onSuccess: async () => {
      toast.success("Post published.");
      setSelectedFile(null);
      selectedFileRef.current = null;
      setUploadedImage(null);
      setForm(initialForm);
      await queryClient.invalidateQueries({ queryKey: ["posts"] });
      router.push(ROUTES.explore);
    },
  });

  function updateForm<K extends keyof FormState>(key: K, value: FormState[K]) {
    setForm((current) => ({ ...current, [key]: value }));
  }

  function resetFileInputs() {
    if (heroFileInputRef.current) {
      heroFileInputRef.current.value = "";
    }

    if (formFileInputRef.current) {
      formFileInputRef.current.value = "";
    }
  }

  function clearSelectedImage() {
    selectedFileRef.current = null;
    setSelectedFile(null);
    setUploadedImage(null);
    uploadMutation.reset();
    publishMutation.reset();
    resetFileInputs();
  }

  async function uploadSelectedFile() {
    const file = selectedFileRef.current;
    if (!file) {
      throw new Error("Choose an image before publishing.");
    }

    const image = await uploadImageApi(file);
    setUploadedImage(image);
    return image;
  }

  function retryUpload() {
    const file = selectedFileRef.current;
    if (!file) {
      toast.error("Choose an image before uploading.");
      return;
    }

    uploadMutation.mutate(file);
  }

  function handleImageChange(file: File | null) {
    uploadMutation.reset();
    publishMutation.reset();

    if (!file) {
      clearSelectedImage();
      return;
    }

    const validationError = getImageValidationError(file);
    if (validationError) {
      toast.error(validationError);
      resetFileInputs();
      return;
    }

    selectedFileRef.current = file;
    setSelectedFile(file);
    setUploadedImage(null);
    uploadMutation.mutate(file);
  }

  const isPublishing = publishMutation.isPending;
  const isUploading = uploadMutation.isPending;
  const isBusy = isPublishing || isUploading;

  return (
    <PageShell
      contentClassName="pb-10"
      topbar={
        <AppTopbar
          actions={[
            {
              id: "explore",
              icon: <Compass className="size-4" />,
              label: "Explore",
              to: ROUTES.explore,
              variant: "outline",
            },
            {
              id: "locations",
              icon: <Map className="size-4" />,
              label: "Locations",
              to: ROUTES.locations,
              variant: "outline",
            },
            {
              id: "back",
              icon: <ArrowLeft className="size-4" />,
              label: "Explore",
              to: ROUTES.explore,
              variant: "outline",
            },
          ]}
          brand="Falzo Upload"
          brandIcon={<Camera className="size-3.5" />}
          mobileMenuTitle="Upload"
          subtitle="Publish an uploaded image through the backend post API."
        />
      }
    >
      {isSessionChecking ? (
        <div className="app-panel flex min-h-80 items-center justify-center rounded-2xl border-[#d6e5f6] bg-white/90">
          <Loader2 className="size-6 animate-spin text-[#2f6fb8]" />
        </div>
      ) : (
        <div className="grid gap-5 lg:grid-cols-[minmax(0,0.92fr)_minmax(360px,0.58fr)]">
          <section className="app-panel overflow-hidden rounded-2xl border-[#d6e5f6] bg-white/92">
            <div className="relative min-h-[420px] bg-[#edf4fb]">
              {previewUrl ? (
                <>
                  <img
                    alt="Selected upload preview"
                    className="h-full min-h-[420px] w-full object-cover"
                    src={previewUrl}
                  />
                  <div className="absolute left-4 right-4 top-4 flex items-center justify-between gap-3">
                    <div className="min-w-0 rounded-full bg-white/90 px-3 py-1.5 text-xs font-semibold text-[#315578] shadow-sm backdrop-blur-xl">
                      <span className="block truncate">
                        {selectedFile?.name}
                      </span>
                    </div>
                    <div className="flex shrink-0 items-center gap-2">
                      <input
                        accept="image/jpeg,image/png,image/webp,image/jpg,image/svg+xml,image/gif"
                        className="sr-only"
                        disabled={isBusy}
                        onChange={(event) => {
                          handleImageChange(event.target.files?.[0] ?? null);
                        }}
                        ref={heroFileInputRef}
                        type="file"
                      />
                      <Button
                        className="rounded-full bg-white/90 shadow-sm backdrop-blur-xl"
                        disabled={isBusy}
                        onClick={() => heroFileInputRef.current?.click()}
                        size="sm"
                        type="button"
                        variant="outline"
                      >
                        <ImagePlus className="size-4" />
                        Change
                      </Button>
                      <Button
                        aria-label="Remove image"
                        className="rounded-full bg-white/90 shadow-sm backdrop-blur-xl"
                        disabled={isBusy}
                        onClick={clearSelectedImage}
                        size="icon-sm"
                        type="button"
                        variant="outline"
                      >
                        <Trash2 className="size-4" />
                      </Button>
                    </div>
                  </div>
                </>
              ) : (
                <label className="flex min-h-[420px] cursor-pointer flex-col items-center justify-center gap-4 px-6 text-center">
                  <span className="inline-flex size-16 items-center justify-center rounded-2xl bg-white text-[#2f6fb8] shadow-[0_16px_38px_-28px_rgb(28_77_128/0.7)]">
                    <ImagePlus className="size-7" />
                  </span>
                  <span className="max-w-sm text-sm font-semibold text-[#315578]">
                    Choose a JPG, PNG, or WebP, SVG, or GIF image to upload.
                  </span>
                  <input
                    accept="image/jpeg,image/png,image/webp,image/jpg,image/svg+xml,image/gif"
                    className="sr-only"
                    disabled={isBusy}
                    onChange={(event) => {
                      handleImageChange(event.target.files?.[0] ?? null);
                    }}
                    ref={heroFileInputRef}
                    type="file"
                  />
                </label>
              )}
            </div>
          </section>

          <form
            className="app-panel space-y-5 rounded-2xl border-[#d6e5f6] bg-white/95 p-5 sm:p-6"
            onSubmit={(event) => {
              event.preventDefault();
              publishMutation.mutate();
            }}
          >
            <div className="space-y-1">
              <p className="text-xs font-semibold uppercase tracking-wide text-[#6485ab]">
                New post
              </p>
              <h1 className="text-2xl font-semibold tracking-normal text-[#15365a]">
                Upload image and publish
              </h1>
            </div>

            <div className="space-y-2">
              <Label htmlFor="image-file">Image</Label>
              <Input
                accept="image/jpeg,image/png,image/webp,image/jpg,image/svg+xml,image/gif"
                disabled={isBusy}
                id="image-file"
                onChange={(event) => {
                  handleImageChange(event.target.files?.[0] ?? null);
                }}
                ref={formFileInputRef}
                type="file"
              />
              {selectedFile ? (
                <p className="text-xs font-medium text-[#5f7f9f]">
                  {selectedFile.name} - {formatFileSize(selectedFile.size)}
                </p>
              ) : null}
              {isUploading ? (
                <p className="inline-flex items-center gap-2 text-xs font-semibold text-[#356792]">
                  <Loader2 className="size-3.5 animate-spin" />
                  Uploading image
                </p>
              ) : null}
              {uploadedImage ? (
                <p className="inline-flex items-center gap-2 text-xs font-semibold text-[#20774b]">
                  <Check className="size-3.5" />
                  Image uploaded
                </p>
              ) : null}
              {uploadMutation.isError ? (
                <div className="flex flex-wrap items-center gap-2">
                  <p className="text-xs font-semibold text-[#b42318]">
                    {getApiErrorMessage(uploadMutation.error)}
                  </p>
                  <Button
                    className="h-8 rounded-full"
                    disabled={isPublishing || !selectedFile}
                    onClick={retryUpload}
                    size="xs"
                    type="button"
                    variant="outline"
                  >
                    <RefreshCcw className="size-3.5" />
                    Retry
                  </Button>
                </div>
              ) : null}
            </div>

            <div className="space-y-2">
              <Label htmlFor="caption">Caption</Label>
              <textarea
                className="min-h-32 w-full resize-none rounded-lg border border-input bg-white/90 px-3.5 py-3 text-sm text-foreground shadow-[0_1px_2px_rgb(12_31_56/0.03)] outline-none transition focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/40"
                disabled={isPublishing}
                id="caption"
                maxLength={2000}
                onChange={(event) => updateForm("caption", event.target.value)}
                placeholder="Write a short caption"
                value={form.caption}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="location">Location</Label>
              <Input
                disabled={isPublishing}
                id="location"
                maxLength={255}
                onChange={(event) =>
                  updateForm("locationName", event.target.value)
                }
                placeholder="Ho Chi Minh City"
                value={form.locationName}
              />
            </div>

            <div className="grid gap-3 sm:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="latitude">Latitude</Label>
                <Input
                  disabled={isPublishing}
                  id="latitude"
                  inputMode="decimal"
                  onChange={(event) =>
                    updateForm("latitude", event.target.value)
                  }
                  value={form.latitude}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="longitude">Longitude</Label>
                <Input
                  disabled={isPublishing}
                  id="longitude"
                  inputMode="decimal"
                  onChange={(event) =>
                    updateForm("longitude", event.target.value)
                  }
                  value={form.longitude}
                />
              </div>
            </div>

            <div className="app-panel-soft flex items-center gap-3 rounded-xl border-[#d7e5f4] bg-[#f7fbff] px-4 py-3 text-sm text-[#385c80]">
              <MapPin className="size-4 shrink-0 text-[#2f6fb8]" />
              <span className="min-w-0">
                Location fields are sent directly to the backend post API.
              </span>
            </div>

            <Button
              className="w-full"
              disabled={isBusy || !selectedFile}
              type="submit"
              variant="gradient"
            >
              {isPublishing ? (
                <Loader2 className="size-4 animate-spin" />
              ) : (
                <Upload className="size-4" />
              )}
              Publish
            </Button>

            {publishMutation.isSuccess ? (
              <div className="flex items-center gap-2 text-sm font-semibold text-[#20774b]">
                <Check className="size-4" />
                Published
              </div>
            ) : null}
          </form>
        </div>
      )}
    </PageShell>
  );
}
