// https://www.npmjs.com/package/esbuild-plugin-tailwindcss/v/1.0.3
import esbuild from "esbuild"
import { tailwindPlugin } from "esbuild-plugin-tailwindcss"
import resetDist from "./plugins/reset"
import htmlCreation from "./plugins/html"

const isProduction = process.env.NODE_ENV === "production"

export const buildOptions: esbuild.BuildOptions = {
    entryPoints: [
        "src/index.tsx",
        "src/App.tsx",

        // pages
        "src/pages/TeacherPage.tsx",
        "src/components/manual-grading/Comment.tsx",
        "src/components/Card.tsx",

        // components
        "src/overmind/index.tsx",
        "src/overmind/effects.tsx",
        "src/overmind/state.tsx",
        "src/overmind/internalActions.tsx",
    ],
    write: false, // we will write the files ourselves after the html creation
    outdir: "dist",
    entryNames: "[name]-[hash]",
    bundle: true,
    treeShaking: true,
    splitting: true,
    format: "esm",
    minify: true,
    minifyIdentifiers: true,
    minifySyntax: true,
    minifyWhitespace: true,
    logLevel: isProduction ? "error" : "info",
    sourcemap: "linked",
    tsconfig: "tsconfig.json",
    loader: {
        ".scss": "css",
    },
    define: {
        "process.env.NODE_ENV": JSON.stringify("development"),
    },
    plugins: [
        tailwindPlugin(),
        resetDist,
        htmlCreation
    ]
}
