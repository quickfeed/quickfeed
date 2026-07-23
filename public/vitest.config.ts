import { defineConfig } from "vitest/config"

export default defineConfig({
    test: {
        environment: "jsdom",
        setupFiles: ["./src/__tests__/setup.ts"],
        globals: true,
    },
    resolve: {
        // Include 'module' condition so Vite's SSR resolution picks the ESM build
        // of packages like react-router, preventing a dual-module split that would
        // break Router context when react-router-dom loads its ESM variant.
        conditions: ["module", "import", "node", "default"],
    },
})
