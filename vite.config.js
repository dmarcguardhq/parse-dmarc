import vue from "@vitejs/plugin-vue";
import path from "path";
import { fileURLToPath } from "url";
import { defineConfig } from "vite";
import compression from "vite-plugin-compression2";
import { createHtmlPlugin } from "vite-plugin-html";

var __dirname = path.dirname(fileURLToPath(import.meta.url));

export default defineConfig({
  publicDir: path.resolve(__dirname, "./assets"),
  base: "./",
  plugins: [
    vue(),
    createHtmlPlugin({
      minify: true,
    }),
    compression({
      algorithm: "brotliCompress",
      exclude: [/\.(br)$/, /\.(gz)$/],
      threshold: 1024,
    }),
  ],
  build: {
    outDir: "dist",
    emptyOutDir: true,
    minify: "oxc",
    cssMinify: "lightningcss",
    cssCodeSplit: true,
    reportCompressedSize: true,
    rollupOptions: {
      output: {
        manualChunks: function manualChunks(id) {
          if (id.includes("node_modules")) {
            return "vendor";
          }
        },
      },
    },
    chunkSizeWarningLimit: 1000,
    assetsInlineLimit: 4096,
  },
  server: {
    port: 3000,
    host: "::1",
    proxy: {
      "/api": {
        target: "http://localhost:8080",
        changeOrigin: true,
      },
    },
  },
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  css: {
    transformer: "lightningcss",
  },
});
