import { Navigate, createBrowserRouter } from "react-router-dom";
import { HomePage } from "@/pages/home-page";
import { LoginPage } from "@/pages/login-page";
import { RegisterPage } from "@/pages/register-page";

export const router = createBrowserRouter([
  {
    path: "/",
    element: <HomePage />,
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
    path: "*",
    element: <Navigate replace to="/login" />,
  },
]);
