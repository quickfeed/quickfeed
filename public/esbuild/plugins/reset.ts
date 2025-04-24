import { Plugin } from "esbuild"
import { rmSync, readdirSync } from "fs"

const resetDist: Plugin = {
    name: "Reset Dist Folder",
    setup(build) {
        build.onStart(() => {
            readdirSync("dist").forEach(file => rmSync(`dist/${file}`))
        })
    }
}
export default resetDist
