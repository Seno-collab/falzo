export const ROUTES = {
  home: "/",
  login: "/login",
  register: "/register",
  dashboard: "/dashboard",
  profile: "/profile",
} as const;

export function getDashboardOrRegisterRoute(authenticated: boolean) {
  return authenticated ? ROUTES.dashboard : ROUTES.register;
}
