export const ROUTES = {
  home: "/",
  login: "/login",
  register: "/register",
  dashboard: "/dashboard",
  explore: "/",
  locations: "/locations",
  upload: "/upload",
  profile: "/profile",
} as const;

export function getDashboardOrRegisterRoute(authenticated: boolean) {
  return authenticated ? ROUTES.dashboard : ROUTES.register;
}
