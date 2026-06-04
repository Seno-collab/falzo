"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  ArrowLeft,
  Camera,
  Check,
  Compass,
  Crosshair,
  ImagePlus,
  Loader2,
  Map as MapIcon,
  MapPin,
  RefreshCcw,
  Search,
  Tags,
  Trash2,
  Upload,
} from "lucide-react";
import { useRouter } from "next/navigation";
import { useEffect, useMemo, useRef, useState } from "react";
import { toast } from "sonner";
import { AppTopbar } from "@/components/layout/app-topbar";
import { PageShell } from "@/components/layout/page-shell";
import MapClient, { type Coordinates, type MapPoint } from "@/components/map";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  clearAuthSession,
  getApiErrorMessage,
  getMeApi,
  hasAuthSession,
} from "@/features/auth/api";
import {
  defaultLocationSearch,
  normalizeLocationSearchQuery,
  searchLocationsWithFallbackApi,
} from "@/features/locations/search";
import { getCategoriesApi } from "@/features/categories/api";
import type { Category } from "@/features/categories/types";
import type { Location } from "@/features/locations/types";
import {
  checkImageApi,
  createPostApi,
  uploadImageApi,
} from "@/features/posts/api";
import type { UploadedImage } from "@/features/posts/types";
import type { AppMessages } from "@/i18n/messages";
import { useI18n } from "@/i18n/locale-provider";
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
  latitude: "",
  longitude: "",
};
const currentLocationId = "__current-location";
const mapSelectionLocationId = "__map-selection";
const maxImageSize = 10 * 1024 * 1024;
const maxPostImages = 10;
const acceptedImageTypes = new Set([
  "image/jpeg",
  "image/png",
  "image/webp",
]);
type UploadPageCopy = AppMessages["uploadPage"];

function formatCopy(
  template: string,
  values: Record<string, string | number>,
) {
  return Object.entries(values).reduce(
    (text, [key, value]) => text.replaceAll(`{${key}}`, String(value)),
    template,
  );
}

function parseCoordinate(value: string) {
  const normalized = Number(value);
  return Number.isFinite(normalized) ? normalized : Number.NaN;
}

function isSameFileList(left: File[], right: File[]) {
  return (
    left.length === right.length &&
    left.every((file, index) => {
      const other = right[index];
      return (
        other &&
        file.name === other.name &&
        file.size === other.size &&
        file.lastModified === other.lastModified
      );
    })
  );
}

function formatFileSize(size: number) {
  if (size < 1024 * 1024) {
    return `${Math.ceil(size / 1024)} KB`;
  }

  return `${(size / 1024 / 1024).toFixed(1)} MB`;
}

function getImageUnit(copy: UploadPageCopy, count: number) {
  return count === 1 ? copy.imageUnit : copy.imagesUnit;
}

function getImageValidationError(file: File, copy: UploadPageCopy) {
  if (!acceptedImageTypes.has(file.type)) {
    return copy.unsupportedImageType;
  }

  if (file.size > maxImageSize) {
    return formatCopy(copy.imageTooLarge, {
      size: formatFileSize(maxImageSize),
    });
  }

  return null;
}

function locationToMapPoint(location: Location): MapPoint {
  return {
    id: location.id,
    name: location.name,
    address: location.address,
    latitude: location.latitude,
    longitude: location.longitude,
  };
}

export function UploadImageScreen() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const { locale, messages } = useI18n();
  const copy = messages.uploadPage;
  const selectedFilesRef = useRef<File[]>([]);
  const heroFileInputRef = useRef<HTMLInputElement | null>(null);
  const formFileInputRef = useRef<HTMLInputElement | null>(null);
  const [isSessionChecking, setIsSessionChecking] = useState(true);
  const [selectedFiles, setSelectedFiles] = useState<File[]>([]);
  const [uploadedImages, setUploadedImages] = useState<UploadedImage[]>([]);
  const [form, setForm] = useState<FormState>(initialForm);
  const [currentPosition, setCurrentPosition] = useState<Coordinates | null>(
    null,
  );
  const [isLocating, setIsLocating] = useState(false);
  const [locationSearchInput, setLocationSearchInput] = useState("");
  const [selectedLocation, setSelectedLocation] = useState<Location | null>(
    null,
  );
  const [selectedCategoryIds, setSelectedCategoryIds] = useState<number[]>([]);
  const [submittedLocationSearch, setSubmittedLocationSearch] = useState(
    defaultLocationSearch,
  );

  const previewUrls = useMemo(
    () => selectedFiles.map((file) => URL.createObjectURL(file)),
    [selectedFiles],
  );
  const primaryPreviewUrl = previewUrls[0] ?? null;

  useEffect(() => {
    document.title = copy.documentTitle;
  }, [copy.documentTitle]);

  useEffect(() => {
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
      for (const previewUrl of previewUrls) {
        URL.revokeObjectURL(previewUrl);
      }
    };
  }, [previewUrls]);

  const uploadMutation = useMutation({
    mutationFn: async (files: File[]) => {
      return Promise.all(
        files.map(async (file) => {
          await checkImageApi(file);
          return uploadImageApi(file);
        }),
      );
    },
    onError: (error, files) => {
      if (isSameFileList(selectedFilesRef.current, files)) {
        setUploadedImages([]);
      }
      toast.error(getApiErrorMessage(error));
    },
    onSuccess: (images, files) => {
      if (!isSameFileList(selectedFilesRef.current, files)) {
        return;
      }

      setUploadedImages(images);
      toast.success(images.length > 1 ? copy.imagesUploaded : copy.imageUploaded);
    },
  });

  const locationQuery = useQuery({
    enabled: submittedLocationSearch.trim().length > 0,
    queryKey: ["locations", "upload-search", submittedLocationSearch, locale],
    queryFn: ({ signal }) =>
      searchLocationsWithFallbackApi(submittedLocationSearch, signal, locale),
  });

  const categoriesQuery = useQuery({
    queryKey: ["categories", locale],
    queryFn: ({ signal }) => getCategoriesApi({ signal }),
    staleTime: 5 * 60_000,
  });

  const categories = useMemo<Category[]>(
    () => categoriesQuery.data ?? [],
    [categoriesQuery.data],
  );
  const selectedCategories = useMemo(
    () =>
      categories.filter((category) => selectedCategoryIds.includes(category.id)),
    [categories, selectedCategoryIds],
  );

  const mapPoints = useMemo<MapPoint[]>(() => {
    const points = new Map<string, MapPoint>();

    for (const location of locationQuery.data ?? []) {
      points.set(location.id, locationToMapPoint(location));
    }

    if (selectedLocation && !points.has(selectedLocation.id)) {
      points.set(selectedLocation.id, locationToMapPoint(selectedLocation));
    }

    return Array.from(points.values());
  }, [locationQuery.data, selectedLocation]);

  useEffect(() => {
    if (locationQuery.error) {
      toast.error(getApiErrorMessage(locationQuery.error));
    }
  }, [locationQuery.error]);

  useEffect(() => {
    if (categoriesQuery.error) {
      toast.error(getApiErrorMessage(categoriesQuery.error));
    }
  }, [categoriesQuery.error]);

  const publishMutation = useMutation({
    mutationFn: async () => {
      if (selectedFiles.length === 0 && uploadedImages.length === 0) {
        throw new Error(copy.chooseImageBeforePublishing);
      }

      const images =
        uploadedImages.length > 0 ? uploadedImages : await uploadSelectedFiles();
      const imageUrls = images.map((image) => image.url).filter(Boolean);
      if (imageUrls.length === 0) {
        throw new Error(copy.chooseImageBeforePublishing);
      }
      const latitude = parseCoordinate(form.latitude);
      const longitude = parseCoordinate(form.longitude);
      if (
        !selectedLocation ||
        !form.locationName.trim() ||
        !Number.isFinite(latitude) ||
        !Number.isFinite(longitude)
      ) {
        throw new TypeError(copy.chooseLocationBeforePublishing);
      }
      if (categories.length > 0 && selectedCategories.length === 0) {
        throw new TypeError(copy.chooseCategoryBeforePublishing);
      }

      return createPostApi({
        image_url: imageUrls[0],
        image_urls: imageUrls,
        caption: form.caption,
        ...(selectedCategories.length > 0
          ? { category_ids: selectedCategories.map((category) => category.id) }
          : {}),
        location_name: form.locationName,
        latitude,
        longitude,
      });
    },
    onError: (error) => {
      toast.error(getApiErrorMessage(error));
    },
    onSuccess: async () => {
      toast.success(copy.postPublished);
      setSelectedFiles([]);
      selectedFilesRef.current = [];
      setUploadedImages([]);
      setForm(initialForm);
      setSelectedLocation(null);
      setSelectedCategoryIds([]);
      setLocationSearchInput("");
      await queryClient.invalidateQueries({ queryKey: ["posts"] });
    },
  });

  function updateForm<K extends keyof FormState>(key: K, value: FormState[K]) {
    setForm((current) => ({ ...current, [key]: value }));
  }

  function submitLocationSearch() {
    setSubmittedLocationSearch(
      normalizeLocationSearchQuery(locationSearchInput),
    );
  }

  function selectLocation(location: Location) {
    setSelectedLocation(location);
    setLocationSearchInput(location.name);
    setForm((current) => ({
      ...current,
      locationName: location.name,
      latitude: String(location.latitude),
      longitude: String(location.longitude),
    }));
  }

  function selectCurrentPosition(position: Coordinates) {
    const currentLocation: Location = {
      id: currentLocationId,
      name: copy.currentLocation,
      address: `${position.latitude.toFixed(5)}, ${position.longitude.toFixed(
        5,
      )}`,
      latitude: position.latitude,
      longitude: position.longitude,
    };

    setCurrentPosition(position);
    selectLocation(currentLocation);
  }

  function selectMapLocation(position: Coordinates) {
    const locationName =
      locationSearchInput.trim() ||
      formatCopy(copy.mapLocation, {
        latitude: position.latitude.toFixed(5),
        longitude: position.longitude.toFixed(5),
      });
    const mapLocation: Location = {
      id: mapSelectionLocationId,
      name: locationName,
      address: `${position.latitude.toFixed(5)}, ${position.longitude.toFixed(
        5,
      )}`,
      latitude: position.latitude,
      longitude: position.longitude,
    };

    selectLocation(mapLocation);
  }

  function handleUseCurrentLocation() {
    if (!navigator.geolocation) {
      toast.error(copy.noLocationAccess);
      return;
    }

    setIsLocating(true);
    navigator.geolocation.getCurrentPosition(
      (position) => {
        selectCurrentPosition({
          latitude: position.coords.latitude,
          longitude: position.coords.longitude,
        });
        setIsLocating(false);
      },
      (error) => {
        toast.error(error.message || copy.unableToReadLocation);
        setIsLocating(false);
      },
      {
        enableHighAccuracy: true,
        maximumAge: 30_000,
        timeout: 12_000,
      },
    );
  }

  function resetFileInputs() {
    if (heroFileInputRef.current) {
      heroFileInputRef.current.value = "";
    }

    if (formFileInputRef.current) {
      formFileInputRef.current.value = "";
    }
  }

  function clearSelectedImages() {
    selectedFilesRef.current = [];
    setSelectedFiles([]);
    setUploadedImages([]);
    uploadMutation.reset();
    publishMutation.reset();
    resetFileInputs();
  }

  async function uploadSelectedFiles() {
    const files = selectedFilesRef.current;
    if (files.length === 0) {
      throw new Error(copy.chooseImageBeforePublishing);
    }

    const images = await Promise.all(
      files.map(async (file) => {
        await checkImageApi(file);
        return uploadImageApi(file);
      }),
    );
    setUploadedImages(images);
    return images;
  }

  function toggleCategory(categoryId: number) {
    setSelectedCategoryIds((current) =>
      current.includes(categoryId)
        ? current.filter((id) => id !== categoryId)
        : [...current, categoryId],
    );
  }

  function retryUpload() {
    const files = selectedFilesRef.current;
    if (files.length === 0) {
      toast.error(copy.chooseImageBeforeUploading);
      return;
    }

    uploadMutation.mutate(files);
  }

  function handleImageChange(files: FileList | File[] | null) {
    uploadMutation.reset();
    publishMutation.reset();

    const nextFiles = Array.from(files ?? []);
    if (nextFiles.length === 0) {
      clearSelectedImages();
      return;
    }

    if (nextFiles.length > maxPostImages) {
      toast.error(formatCopy(copy.chooseUpToImages, { count: maxPostImages }));
      resetFileInputs();
      return;
    }

    const validationError = nextFiles
      .map((file) => getImageValidationError(file, copy))
      .find(Boolean);
    if (validationError) {
      toast.error(validationError);
      resetFileInputs();
      return;
    }

    selectedFilesRef.current = nextFiles;
    setSelectedFiles(nextFiles);
    setUploadedImages([]);
    uploadMutation.mutate(nextFiles);
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
              label: copy.navExplore,
              to: ROUTES.explore,
              variant: "outline",
            },
            {
              id: "locations",
              icon: <MapIcon className="size-4" />,
              label: copy.navLocations,
              to: ROUTES.locations,
              variant: "outline",
            },
            {
              id: "back",
              icon: <ArrowLeft className="size-4" />,
              label: copy.navExplore,
              to: ROUTES.explore,
              variant: "outline",
            },
          ]}
          brand={copy.brand}
          brandIcon={<Camera className="size-3.5" />}
          mobileMenuTitle={copy.mobileMenuTitle}
          subtitle={copy.topbarSubtitle}
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
            <div className="relative min-h-105 bg-[#edf4fb]">
              {primaryPreviewUrl ? (
                <>
                  <img
                    alt={copy.selectedUploadPreview}
                    className="h-full min-h-105 w-full object-cover"
                    decoding="async"
                    src={primaryPreviewUrl}
                  />
                  <div className="absolute left-4 right-4 top-4 flex items-center justify-between gap-3">
                    <div className="min-w-0 rounded-full bg-white/90 px-3 py-1.5 text-xs font-semibold text-[#315578] shadow-sm backdrop-blur-xl">
                      <span className="block truncate">
                        {selectedFiles.length === 1
                          ? selectedFiles[0]?.name
                          : formatCopy(copy.imagesSelected, {
                              count: selectedFiles.length,
                            })}
                      </span>
                    </div>
                    <div className="flex shrink-0 items-center gap-2">
                      <input
                        accept="image/jpeg,image/png,image/webp"
                        className="sr-only"
                        disabled={isBusy}
                        onChange={(event) => {
                          handleImageChange(event.target.files);
                        }}
                        multiple
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
                        {copy.change}
                      </Button>
                      <Button
                        aria-label={copy.removeImages}
                        className="rounded-full bg-white/90 shadow-sm backdrop-blur-xl"
                        disabled={isBusy}
                        onClick={clearSelectedImages}
                        size="icon-sm"
                        type="button"
                        variant="outline"
                      >
                        <Trash2 className="size-4" />
                      </Button>
                    </div>
                  </div>
                  {previewUrls.length > 1 ? (
                    <div className="absolute bottom-4 left-4 right-4 flex gap-2 overflow-x-auto pb-1 [&::-webkit-scrollbar]:hidden">
                      {previewUrls.map((url, index) => (
                        <span
                          className="relative h-16 w-16 shrink-0 overflow-hidden rounded-xl border-2 border-white bg-white shadow-lg"
                          key={`${selectedFiles[index]?.name ?? "image"}-${index}`}
                        >
                          <img
                            alt={formatCopy(copy.selectedUploadAlt, {
                              index: index + 1,
                            })}
                            className="h-full w-full object-cover"
                            decoding="async"
                            src={url}
                          />
                        </span>
                      ))}
                    </div>
                  ) : null}
                </>
              ) : (
                <label className="flex min-h-105 cursor-pointer flex-col items-center justify-center gap-4 px-6 text-center">
                  <span className="inline-flex size-16 items-center justify-center rounded-2xl bg-white text-[#2f6fb8] shadow-[0_16px_38px_-28px_rgb(28_77_128/0.7)]">
                    <ImagePlus className="size-7" />
                  </span>
                  <span className="max-w-sm text-sm font-semibold text-[#315578]">
                    {formatCopy(copy.chooseImagesPrompt, {
                      count: maxPostImages,
                    })}
                  </span>
                  <input
                    accept="image/jpeg,image/png,image/webp"
                    className="sr-only"
                    disabled={isBusy}
                    onChange={(event) => {
                      handleImageChange(event.target.files);
                    }}
                    multiple
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
                {copy.newPost}
              </p>
              <h1 className="text-2xl font-semibold tracking-normal text-[#15365a]">
                {copy.title}
              </h1>
            </div>

            <div className="space-y-2">
              <Label htmlFor="image-file">{copy.imageLabel}</Label>
              <Input
                accept="image/jpeg,image/png,image/webp"
                disabled={isBusy}
                id="image-file"
                onChange={(event) => {
                  handleImageChange(event.target.files);
                }}
                multiple
                ref={formFileInputRef}
                type="file"
              />
              {selectedFiles.length > 0 ? (
                <p className="text-xs font-medium text-[#5f7f9f]">
                  {formatCopy(copy.selectedImagesSummary, {
                    count: selectedFiles.length,
                    unit: getImageUnit(copy, selectedFiles.length),
                    size: formatFileSize(
                      selectedFiles.reduce((total, file) => total + file.size, 0),
                    ),
                  })}
                </p>
              ) : null}
              {isUploading ? (
                <p className="inline-flex items-center gap-2 text-xs font-semibold text-[#356792]">
                  <Loader2 className="size-3.5 animate-spin" />
                  {formatCopy(copy.uploadingImages, {
                    unit: getImageUnit(copy, selectedFiles.length),
                  })}
                </p>
              ) : null}
              {uploadedImages.length > 0 ? (
                <p className="inline-flex items-center gap-2 text-xs font-semibold text-[#20774b]">
                  <Check className="size-3.5" />
                  {formatCopy(copy.uploadedImages, {
                    count: uploadedImages.length,
                    unit: getImageUnit(copy, uploadedImages.length),
                  })}
                </p>
              ) : null}
              {uploadMutation.isError ? (
                <div className="flex flex-wrap items-center gap-2">
                  <p className="text-xs font-semibold text-[#b42318]">
                    {getApiErrorMessage(uploadMutation.error)}
                  </p>
                  <Button
                    className="h-8 rounded-full"
                    disabled={isPublishing || selectedFiles.length === 0}
                    onClick={retryUpload}
                    size="xs"
                    type="button"
                    variant="outline"
                  >
                    <RefreshCcw className="size-3.5" />
                    {copy.retry}
                  </Button>
                </div>
              ) : null}
            </div>

            <div className="space-y-2">
              <Label htmlFor="caption">{copy.captionLabel}</Label>
              <textarea
                className="min-h-32 w-full resize-none rounded-lg border border-input bg-white/90 px-3.5 py-3 text-sm text-foreground shadow-[0_1px_2px_rgb(12_31_56/0.03)] outline-none transition focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/40"
                disabled={isPublishing}
                id="caption"
                maxLength={2000}
                onChange={(event) => updateForm("caption", event.target.value)}
                placeholder={copy.captionPlaceholder}
                value={form.caption}
              />
            </div>

            <div className="space-y-3">
              <div className="flex items-center justify-between gap-3">
                <Label>{copy.categoriesLabel}</Label>
                {categoriesQuery.isFetching ? (
                  <span className="inline-flex items-center gap-1.5 text-xs font-semibold text-[#356792]">
                    <Loader2 className="size-3.5 animate-spin" />
                    {copy.loading}
                  </span>
                ) : null}
              </div>

              {categories.length > 0 ? (
                <div className="flex max-h-32 flex-wrap gap-2 overflow-y-auto pr-1">
                  {categories.map((category) => {
                    const selected = selectedCategoryIds.includes(category.id);

                    return (
                      <button
                        aria-pressed={selected}
                        className={`inline-flex items-center gap-2 rounded-full border px-3 py-2 text-sm font-semibold transition ${
                          selected
                            ? "border-[#2f6fb8] bg-[#15365a] text-white"
                            : "border-[#d7e5f4] bg-white/90 text-[#385c80] hover:border-[#a9c8e8] hover:bg-[#f8fbff]"
                        }`}
                        disabled={isPublishing}
                        key={category.id}
                        onClick={() => toggleCategory(category.id)}
                        type="button"
                      >
                        {selected ? (
                          <Check className="size-3.5" />
                        ) : (
                          <Tags className="size-3.5" />
                        )}
                        {category.name}
                      </button>
                    );
                  })}
                </div>
              ) : (
                <div className="app-panel-soft flex items-center gap-3 rounded-xl border-[#d7e5f4] bg-[#f7fbff] px-4 py-3 text-sm text-[#385c80]">
                  <Tags className="size-4 shrink-0 text-[#2f6fb8]" />
                  <span className="min-w-0">
                    {copy.noCategories}
                  </span>
                </div>
              )}
            </div>

            <div className="space-y-3">
              <Label htmlFor="location-search">{copy.locationLabel}</Label>
              <div className="grid gap-2 sm:grid-cols-[minmax(0,1fr)_auto_auto]">
                <Input
                  disabled={isPublishing}
                  id="location-search"
                  maxLength={255}
                  onChange={(event) =>
                    setLocationSearchInput(event.target.value)
                  }
                  onKeyDown={(event) => {
                    if (event.key === "Enter") {
                      event.preventDefault();
                      submitLocationSearch();
                    }
                  }}
                  placeholder={copy.locationPlaceholder}
                  value={locationSearchInput}
                />
                <Button
                  disabled={isPublishing || locationQuery.isFetching}
                  onClick={submitLocationSearch}
                  type="button"
                  variant="outline"
                >
                  {locationQuery.isFetching ? (
                    <Loader2 className="size-4 animate-spin" />
                  ) : (
                    <Search className="size-4" />
                  )}
                  {copy.search}
                </Button>
                <Button
                  disabled={isPublishing || isLocating}
                  onClick={handleUseCurrentLocation}
                  type="button"
                  variant="outline"
                >
                  {isLocating ? (
                    <Loader2 className="size-4 animate-spin" />
                  ) : (
                    <Crosshair className="size-4" />
                  )}
                  {copy.current}
                </Button>
              </div>

              <MapClient
                currentPosition={currentPosition}
                height="compact"
                onSelectCoordinates={selectMapLocation}
                onSelectPoint={(point) => {
                  const nextLocation =
                    locationQuery.data?.find(
                      (location) => location.id === point.id,
                    ) ??
                    (selectedLocation?.id === point.id
                      ? selectedLocation
                      : null);

                  if (nextLocation) {
                    selectLocation(nextLocation);
                  }
                }}
                points={mapPoints}
                selectedPointId={selectedLocation?.id}
                zoom={13}
              />

              <div className="grid max-h-52 gap-2 overflow-y-auto pr-1">
                {locationQuery.data?.map((location) => {
                  const selected = selectedLocation?.id === location.id;

                  return (
                    <button
                      className={`rounded-xl border px-3 py-2 text-left transition ${
                        selected
                          ? "border-[#2f6fb8] bg-[#eef7ff]"
                          : "border-[#d7e5f4] bg-white/90 hover:border-[#a9c8e8] hover:bg-[#f8fbff]"
                      }`}
                      disabled={isPublishing}
                      key={location.id}
                      onClick={() => selectLocation(location)}
                      type="button"
                    >
                      <span className="block truncate text-sm font-semibold text-[#15365a]">
                        {location.name}
                      </span>
                      <span className="mt-1 line-clamp-2 text-xs leading-5 text-[#6682a1]">
                        {location.address}
                      </span>
                    </button>
                  );
                })}

                {!locationQuery.isFetching &&
                locationQuery.data?.length === 0 ? (
                  <div className="rounded-xl border border-[#d7e5f4] bg-white/90 px-3 py-3 text-sm font-semibold text-[#6682a1]">
                    {copy.noLocations}
                  </div>
                ) : null}
              </div>

              {selectedLocation ? (
                <div className="app-panel-soft flex items-start gap-3 rounded-xl border-[#d7e5f4] bg-[#f7fbff] px-4 py-3 text-sm text-[#385c80]">
                  <MapPin className="mt-0.5 size-4 shrink-0 text-[#2f6fb8]" />
                  <span className="min-w-0">
                    <span className="block font-semibold text-[#15365a]">
                      {selectedLocation.name}
                    </span>
                    <span className="block text-xs leading-5 text-[#6682a1]">
                      {selectedLocation.address}
                    </span>
                  </span>
                </div>
              ) : (
                <div className="app-panel-soft flex items-center gap-3 rounded-xl border-[#d7e5f4] bg-[#f7fbff] px-4 py-3 text-sm text-[#385c80]">
                  <MapPin className="size-4 shrink-0 text-[#2f6fb8]" />
                  <span className="min-w-0">
                    {copy.chooseLocationHint}
                  </span>
                </div>
              )}
            </div>

            <Button
              className="w-full"
              disabled={
                isBusy ||
                selectedFiles.length === 0 ||
                !selectedLocation ||
                (categories.length > 0 && selectedCategories.length === 0)
              }
              type="submit"
              variant="gradient"
            >
              {isPublishing ? (
                <Loader2 className="size-4 animate-spin" />
              ) : (
                <Upload className="size-4" />
              )}
              {copy.publish}
            </Button>

            {publishMutation.isSuccess ? (
              <div className="flex items-center gap-2 text-sm font-semibold text-[#20774b]">
                <Check className="size-4" />
                {copy.published}
              </div>
            ) : null}
          </form>
        </div>
      )}
    </PageShell>
  );
}
