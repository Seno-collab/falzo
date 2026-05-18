export const ROUTES = {
  home: "/",
  login: "/login",
  register: "/register",
  dashboard: "/dashboard",
  explore: "/",
  saved: "/saved",
  savedCollection: (shareSlug: string) =>
    `/shared?collection=${encodeURIComponent(shareSlug)}`,
  locations: "/locations",
  upload: "/upload",
  profile: "/profile",
  userProfile: (userId: number | string) =>
    `/users?userId=${encodeURIComponent(String(userId))}`,
} as const;

export function getDashboardOrRegisterRoute(authenticated: boolean) {
  return authenticated ? ROUTES.dashboard : ROUTES.register;
}
