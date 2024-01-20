import { fileURLToPath } from "node:url";
import { mergeConfig, defineConfig, configDefaults } from "vitest/config";
import viteConfig from "./vite.config";

export default mergeConfig(
  viteConfig,
  defineConfig({
    test: {
      globals: true,
      server: {
        deps: {
          inline: ["vuetify"],
        },
      },
      coverage: {
        exclude: ["src/main.ts", ".eslintrc.cjs", "env.d.ts", "**/types.ts", "src/router/index.ts"],
      },
      environment: "jsdom",
      exclude: [...configDefaults.exclude, "e2e/*"],
      root: fileURLToPath(new URL("./", import.meta.url)),
    },
  }),
);
