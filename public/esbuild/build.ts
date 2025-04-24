import { build } from "esbuild"
import { buildOptions } from "./options"

build(buildOptions).then(() => {
    console.log("Build completed successfully")
}).catch((error) => {
    console.error("Build failed:", error)
})
