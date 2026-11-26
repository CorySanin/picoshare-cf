import js from "@eslint/js";
import html from "eslint-plugin-html";
import playwright from "eslint-plugin-playwright";
import globals from "globals";
import tseslint from "typescript-eslint";

export default [
  {
    ignores: [
      "e2e/playwright-report/*",
      "handlers/static/third-party/**/*.js",
      "result*/",
    ],
  },
  {
    files: ["handlers/templates/**/*.html"],
    plugins: { html },
  },
  {
    ...js.configs.recommended,
    languageOptions: {
      ...js.configs.recommended.languageOptions,
      ecmaVersion: 2022,
      sourceType: "module",
      globals: {
        ...globals.browser,
        ...globals.node,
      },
    },
    rules: {
      ...js.configs.recommended.rules,
      "no-console": [
        process.env.NODE_ENV === "production" ? "error" : "warn",
        { allow: ["error"] },
      ],
    },
  },
  ...tseslint.configs.recommended.map((config) => ({
    ...config,
    files: ["e2e/**/*.ts"],
  })),
  {
    ...playwright.configs["flat/recommended"],
    files: ["e2e/**/*.ts"],
    rules: {
      ...playwright.configs["flat/recommended"].rules,
      "playwright/no-networkidle": "off",
    },
  },
];
