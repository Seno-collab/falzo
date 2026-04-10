import React from "react";
import ReactDOM from "react-dom/client";
import App from "@/App";
import { initializeAuthHeader } from "@/api/auth.api";
import "./styles.css";

initializeAuthHeader();

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
