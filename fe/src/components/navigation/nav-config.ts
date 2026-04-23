import { Compass, Heart, Map, PlusSquare, UserRound } from "lucide-react";
import { ROUTES } from "@/lib/routes";

export const navigationItems = [
  {
    href: ROUTES.explore,
    label: "Explore",
    icon: Compass,
  },
  {
    href: ROUTES.map,
    label: "Map",
    icon: Map,
  },
  {
    href: ROUTES.create,
    label: "Create",
    icon: PlusSquare,
  },
  {
    href: ROUTES.saved,
    label: "Saved",
    icon: Heart,
  },
  {
    href: ROUTES.profile,
    label: "Profile",
    icon: UserRound,
  },
] as const;
