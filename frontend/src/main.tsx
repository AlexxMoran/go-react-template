import { App } from "@app/index";
import { createRoot } from "react-dom/client";
import { BrowserRouter } from "react-router";

const root = document.getElementById("root");

createRoot(root!).render(
  <BrowserRouter>
    <App />
  </BrowserRouter>
);
