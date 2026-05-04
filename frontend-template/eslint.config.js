import { defineConfig } from "eslint/config";
import js from "@eslint/js";
import globals from "globals";

export default defineConfig([
  {
    ignores: ["public/**/*.min.js", "node_modules/**"],
  },
  js.configs.recommended,
  {
    files: ["public/scripts/**/*.js"],
    languageOptions: {
      ecmaVersion: 2025,
      sourceType: "module",
      globals: {
        ...globals.browser,
        htmx: "readonly",
      },
    },
    rules: {
      "no-unused-vars": "warn",
      "no-undef": "error",
      "eqeqeq": "error",
      "no-console": "warn",
      "no-var": "error",
      "prefer-const": "error",
    },
  },
]);