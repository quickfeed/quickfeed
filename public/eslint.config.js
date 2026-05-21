// ESLint flat config (ESLint v9+).
// See https://eslint.org/docs/latest/use/configure/configuration-files.

import { defineConfig } from "eslint/config"

import js from "@eslint/js"
import tsParser from "@typescript-eslint/parser"
import tsPlugin from "@typescript-eslint/eslint-plugin"
import reactPlugin from "eslint-plugin-react"
import reactHooksPlugin from "eslint-plugin-react-hooks"

export default defineConfig([
    // Global ignores: per the docs, an `ignores` entry is global only when
    // it is the sole key in its config object. node_modules/ is ignored by
    // ESLint automatically. dist/ is the esbuild output; proto/ is generated.
    // Plain .js/.cjs/.mjs files (eslint.config.js, setup.jest.js, ...) are
    // excluded to mirror the prior `--ext .ts,.tsx` scope.
    { ignores: ["dist/**", "proto/**", "**/*.js", "**/*.cjs", "**/*.mjs", "**/*.jsx"] },

    // Recommended baseline presets. Pushed as full config objects (not by
    // spreading their `rules`) so any `languageOptions`, `plugins`, and
    // `files` they ship with come along too.
    js.configs.recommended,
    tsPlugin.configs["flat/eslint-recommended"],
    reactPlugin.configs.flat.recommended,
    // react-hooks ships two presets: `recommended` is legacy (eslintrc) and
    // `recommended-latest` is the flat-config equivalent.
    reactHooksPlugin.configs["recommended-latest"],

    // TypeScript wiring and project-specific overrides.
    {
        files: ["**/*.ts", "**/*.tsx"],
        languageOptions: {
            parser: tsParser,
            ecmaVersion: 6,
            sourceType: "module",
            parserOptions: {
                ecmaFeatures: { jsx: true },
            },
        },
        plugins: {
            // Registered here so the `@typescript-eslint/no-unused-vars`
            // override below resolves. The preset above doesn't register
            // the plugin itself.
            "@typescript-eslint": tsPlugin,
        },
        settings: {
            // Quiet "React version not specified" warning by letting the
            // plugin auto-detect from package.json.
            react: { version: "detect" },
        },
        rules: {
            // Project rule overrides (carried over verbatim from the prior
            // .eslintrc.json).
            "react/function-component-definition": [2, {
                namedComponents: "arrow-function",
                unnamedComponents: ["function-expression", "arrow-function"],
            }],
            "react/jsx-filename-extension": [2, { extensions: [".ts", ".tsx"] }],
            "react/jsx-newline": "off",
            "no-undef": "off",
            "no-unused-vars": "off",
            "@typescript-eslint/no-unused-vars": "warn",
            "react/jsx-max-depth": [2, { max: 6 }],
            // Disable the no-bind rule
            "react/jsx-no-bind": "off",
            // Prefer self-closing tags for elements without children.
            "react/self-closing-comp": ["warn", { component: true, html: true }],

            // --- Additional project rules ---
            // Push back on `any`; defeats the point of TS.
            "@typescript-eslint/no-explicit-any": "warn",
            // Use `import type { X }` for type-only imports to keep them
            // out of the runtime bundle.
            "@typescript-eslint/consistent-type-imports": "warn",
            // Flag unnecessary fragments like <>{x}</>. Auto-fixable.
            "react/jsx-no-useless-fragment": "warn",
            // Prefer <Foo disabled /> over <Foo disabled={true} />.
            "react/jsx-boolean-value": ["error", "never"],
            // Require === and !==.
            eqeqeq: ["error", "always"],
            // Using array indices as keys breaks reconciliation on reorder.
            "react/no-array-index-key": "warn",
            // Drop unnecessary curly braces: <Foo bar="hello" /> not bar={"hello"}.
            "react/jsx-curly-brace-presence": ["error", { props: "never", children: "never" }],
            // Always require braces around control-flow bodies.
            curly: ["error", "all"],
        },
    },
])
