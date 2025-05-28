import React from "react";
import ReactDOM from "react-dom/client";
import Dashboard from "./pages/Dashboard";
import "./styles/tailwind.css";

// Если потребуется поддержка других страниц/роутинга, можно добавить react-router-dom

const rootElement = document.getElementById("root");

if (rootElement) {
  ReactDOM.createRoot(rootElement).render(
    <React.StrictMode>
      <Dashboard />
    </React.StrictMode>
  );
} else {
  // fallback для старых шаблонов или если id="root" не найден
  const fallback = document.createElement("div");
  fallback.id = "root";
  document.body.appendChild(fallback);
  ReactDOM.createRoot(fallback).render(
    <React.StrictMode>
      <Dashboard />
    </React.StrictMode>
  );
}
