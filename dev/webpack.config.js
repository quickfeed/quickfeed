var webpack = require("webpack")

module.exports = {
    entry: "./src/index.tsx",
    output: {
        filename: "bundle.js",
        path: __dirname + "/dist"
    },
    mode: "development",
    // watch enables webpacks Watch flag, which means it will run endlessly and recompile on saves
    watch: true,
        
    watchOptions: {
        // Poll for file changes every 1000ms. Required for --watch to function in Docker containers
        poll: 1000,
    },
    // Enable sourcemaps for debugging webpack's output.
    devtool: "source-map",

    resolve: {
        // Add '.ts' and '.tsx' as resolvable extensions.
        extensions: [".ts", ".tsx", ".js", ".json"],
    },


    plugins: [
        new webpack.ProvidePlugin({
            process: 'process/browser'
        }),
        new webpack.DefinePlugin({
            'process.env.ASSET_PATH': JSON.stringify("static"),
        }),
    ],

    module: {
        rules: [
            // All files with a '.ts' or '.tsx' extension will be handled by 'ts-loader'.
            { test: /\.tsx?$/, loader: "ts-loader" },

            // All output '.js' files will have any sourcemaps re-processed by 'source-map-loader'.
            { enforce: "pre", test: /\.js$/, loader: "source-map-loader" },
            { test: /\.css$/i,use: ["style-loader", "css-loader"],},
        ]
    },
    // When importing a module whose path matches one of the following, just
    // assume a corresponding global variable exists and use that instead.
    // This is important because it allows us to avoid bundling all of our
    // dependencies, which allows browsers to cache those libraries between builds.
    externals: {
        "react": "React",
        "react-dom": "ReactDOM",
    },
};