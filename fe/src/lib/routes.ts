export const ROUTES = {
  home: "/",
  explore: "/explore",
  login: "/login",
  register: "/register",
  dashboard: "/dashboard",
  scenicGallery: "/scenic-gallery",
  travel3d: "/travel-3d",
  map: "/map",
  create: "/create",
  saved: "/saved",
  profile: "/profile",
} as const;

export function getDashboardOrRegisterRoute(authenticated: boolean) {
  return authenticated ? ROUTES.dashboard : ROUTES.register;
}
