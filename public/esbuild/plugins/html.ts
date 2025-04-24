import { Plugin } from "esbuild"
import { readFile, writeFile } from "fs"
import path from "path"

export const htmlCreation: Plugin = {
    name: "Copy HTML",
    setup(build) {
        build.onEnd(async (result) => {
            readFile("index.tmpl.html", async (err, data) => {
                if (err) console.error("Error reading index.tmpl.html:", err)

                const links: string[] = []
                if (!result.outputFiles) {
                    console.error("No output files found in the result.")
                    return
                }
                result.outputFiles.forEach((element) => {
                    const fileName = path.basename(element.path)
                    const ext = path.extname(fileName)

                    if (ext === ".js") {
                        links.push(`     <script type="module" src="/static/${fileName}" defer></script>`)
                    } else if (ext === ".css") {
                        links.push(`     <link rel="stylesheet" href="/static/${fileName}" />`)
                    }
                })
                const file = data.toString().split("\n")
                const index = file.findIndex((line) => line.includes("</head>"))
                if (index === -1) {
                    console.error("Error: </head> tag not found in index.tmpl.html")
                    return
                }
                file.splice(index, 0, ...links)

                const newFile = file.join("\n")

                writeFile("assets/index.html", newFile, (err) => {
                    if (err) console.error("Error writing index.html:", err)
                })

                if (result.outputFiles) {
                    result.outputFiles.forEach(element => {
                        writeFile(element.path, element.contents, (err) => {
                            if (err) console.error("Error writing file:", err)
                        })
                    })
                } else {
                    console.error("No output files found in the result.")
                }
            })
        })
    }
}

export default htmlCreation
