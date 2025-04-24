import { context } from "esbuild"
import { buildOptions } from "./options"

context(buildOptions).then(async (ctx) => {
    return await ctx.watch()
}).catch((error) => {
    console.error("Watcher failed:", error)
})
