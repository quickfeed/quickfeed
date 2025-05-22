import esbuild from 'esbuild'
import tailwindPlugin from 'esbuild-plugin-tailwindcss'

const isWatch = process.argv.includes('--watch')
const opts = {
    entryPoints: ['src/index.tsx'],
    bundle: true,
    outdir: 'assets',      // ← changed from “dist”
    loader: { '.scss': 'css', '.css': 'css' },
    plugins: [tailwindPlugin()],
}

async function run() {
    if (isWatch) {
        const ctx = await esbuild.context(opts)
        await ctx.watch()
    } else {
        await esbuild.build(opts)
    }
}
run().catch(() => process.exit(1))
