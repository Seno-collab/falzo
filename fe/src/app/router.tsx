import { Navigate, createBrowserRouter } from "react-router-dom";
import { DashboardPage } from "@/pages/dashboard-page";
import { HomePage } from "@/pages/home-page";
import { LoginPage } from "@/pages/login-page";
import { RegisterPage } from "@/pages/register-page";
import { Travel3DPage } from "@/pages/travel-3d-page";

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
    path: "/travel-3d",
    element: <Travel3DPage />,
  },
  {
    path: "*",
    element: <Navigate replace to="/" />,
  },
]);
