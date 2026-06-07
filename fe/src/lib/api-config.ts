export function envEndpoint(
  nextPublicValue: string | undefined,
  viteValue: string | undefined,
  fallback: string,
) {
  return nextPublicValue ?? viteValue ?? fallback;
}

export function trimTrailingSlashes(path: string) {
  return path.replace(/\/+$/, "");
}

export function endpointPath(
  base: string,
  ...segments: Array<number | string>
) {
  const normalizedBase = trimTrailingSlashes(base);
  const suffix = segments
    .map((segment) => String(segment).replaceAll(/^\/+|\/+$/g, ""))
    .filter(Boolean)
    .join("/");

  return suffix ? `${normalizedBase}/${suffix}` : normalizedBase;
}

export const API_BASE_URL = envEndpoint(
  process.env.NEXT_PUBLIC_API_BASE_URL,
  process.env.VITE_API_BASE_URL,
  "/api",
).trim();

export const AUTH_ENDPOINTS = {
  login: envEndpoint(
    process.env.NEXT_PUBLIC_AUTH_LOGIN_ENDPOINT,
    process.env.VITE_AUTH_LOGIN_ENDPOINT,
    "/auth/login",
  ),
  register: envEndpoint(
    process.env.NEXT_PUBLIC_AUTH_REGISTER_ENDPOINT,
    process.env.VITE_AUTH_REGISTER_ENDPOINT,
    "/auth/register",
  ),
  me: envEndpoint(
    process.env.NEXT_PUBLIC_AUTH_ME_ENDPOINT,
    process.env.VITE_AUTH_ME_ENDPOINT,
    "/auth/me",
  ),
  refresh: envEndpoint(
    process.env.NEXT_PUBLIC_AUTH_REFRESH_ENDPOINT,
    process.env.VITE_AUTH_REFRESH_ENDPOINT,
    "/auth/refresh",
  ),
  logout: envEndpoint(
    process.env.NEXT_PUBLIC_AUTH_LOGOUT_ENDPOINT,
    process.env.VITE_AUTH_LOGOUT_ENDPOINT,
    "/auth/logout",
  ),
  changePassword: envEndpoint(
    process.env.NEXT_PUBLIC_AUTH_CHANGE_PASSWORD_ENDPOINT,
    process.env.VITE_AUTH_CHANGE_PASSWORD_ENDPOINT,
    "/auth/change-password",
  ),
  updateAvatar: envEndpoint(
    process.env.NEXT_PUBLIC_AUTH_UPDATE_AVATAR_ENDPOINT,
    process.env.VITE_AUTH_UPDATE_AVATAR_ENDPOINT,
    "/auth/me/avatar",
  ),
} as const;

export const POSTS_ENDPOINT = envEndpoint(
  process.env.NEXT_PUBLIC_POSTS_ENDPOINT,
  process.env.VITE_POSTS_ENDPOINT,
  "/posts/",
);

export const IMAGE_UPLOAD_ENDPOINT = envEndpoint(
  process.env.NEXT_PUBLIC_IMAGE_UPLOAD_ENDPOINT,
  process.env.VITE_IMAGE_UPLOAD_ENDPOINT,
  "/images/upload",
);

export const IMAGE_CHECK_ENDPOINT = envEndpoint(
  process.env.NEXT_PUBLIC_IMAGE_CHECK_ENDPOINT,
  process.env.VITE_IMAGE_CHECK_ENDPOINT,
  "/images/check",
);

export const CATEGORIES_ENDPOINT = envEndpoint(
  process.env.NEXT_PUBLIC_CATEGORIES_ENDPOINT,
  process.env.VITE_CATEGORIES_ENDPOINT,
  "/categories/",
);

export const LOCATIONS_ENDPOINT = envEndpoint(
  process.env.NEXT_PUBLIC_LOCATIONS_ENDPOINT,
  process.env.VITE_LOCATIONS_ENDPOINT,
  "/locations",
);

export const PLACES_ENDPOINT = envEndpoint(
  process.env.NEXT_PUBLIC_PLACES_ENDPOINT,
  process.env.VITE_PLACES_ENDPOINT,
  "/places",
);

export const NOTIFICATIONS_ENDPOINT = envEndpoint(
  process.env.NEXT_PUBLIC_NOTIFICATIONS_ENDPOINT,
  process.env.VITE_NOTIFICATIONS_ENDPOINT,
  "/notifications",
);

export const USERS_ENDPOINT = envEndpoint(
  process.env.NEXT_PUBLIC_USERS_ENDPOINT,
  process.env.VITE_USERS_ENDPOINT,
  "/users",
);

export function buildApiUrl(path: string) {
  if (/^https?:\/\//i.test(path)) {
    return path;
  }

  const normalizedBase = API_BASE_URL.replace(/\/+$/, "");
  const normalizedPath = path.replace(/^\/+/, "");
  if (!normalizedBase || normalizedBase === "/") {
    return `/${normalizedPath}`;
  }

  return `${normalizedBase}/${normalizedPath}`;
}
