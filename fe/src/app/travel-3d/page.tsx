import { redirect } from "next/navigation";
import { ROUTES } from "@/lib/routes";

export default function Travel3DRoutePage() {
  redirect(ROUTES.scenicGallery);
}
