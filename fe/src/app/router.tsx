import { Navigate, createBrowserRouter } from "react-router-dom";
import { DashboardPage } from "@/pages/dashboard-page";
import { HomePage } from "@/pages/home-page";
import { LoginPage } from "@/pages/login-page";
import { RegisterPage } from "@/pages/register-page";
import { ScenicGalleryPage } from "@/pages/scenic-gallery-page";

export const router = createBrowserRouter([
  {
    path: "/",
    element: <HomePage />,
  },
  {
    path: "/dashboard",
    element: <DashboardPage />,
  },
  {
    path: "/login",
    element: <LoginPage />,
  },
  {
    path: "/register",
    element: <RegisterPage />,
  },
  {
    path: "/scenic-gallery",
    element: <ScenicGalleryPage />,
  },
  {
    path: "/travel-3d",
    element: <Navigate replace to="/scenic-gallery" />,
  },
  {
    path: "*",
    element: <Navigate replace to="/" />,
  },
]);
