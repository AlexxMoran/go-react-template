import react from "@vitejs/plugin-react";
import { resolve } from "path";
import { defineConfig } from "vite";

const apiProxyTarget = process.env.VITE_API_PROXY_TARGET || "http://localhost:3000";

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "@app": resolve(__dirname, "src/app"),
      "@pages": resolve(__dirname, "src/pages"),
      "@shared": resolve(__dirname, "src/shared")
    }
  },
  server: {
    proxy: {
      "/api": {
        target: apiProxyTarget,
        changeOrigin: true,
        secure: false
      }
    },
    port: 4200,
    // Auto-open the browser only outside containers (avoids xdg-open errors in Docker).
    open: !process.env.VITE_API_PROXY_TARGET
  }
});
