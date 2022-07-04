const webpack = require("webpack")
const HtmlWebpackPlugin = require("html-webpack-plugin")

module.exports = {
    entry: {
        index: {
            import: "./src/index.tsx",
            dependOn: 'proto',
            dependOn: 'overmind'
        },
        overmind: {
            import: "./src/overmind/index.tsx",
            dependOn: 'proto',
        },
        proto: {
            import: "./proto/ag/ag_pb.js",
            dependOn: "protobuf",
        },
        protobuf: "google-protobuf",
    },
    output: {
        // Bundle filenames include hashes based on the contents of the file.
        // This forces the client to reload the bundle if the file changes.
        filename: "[name].[contenthash].bundle.js",
        path: __dirname + "/dist",
        // clean up the dist folder before building
        clean: true
    },
    mode: "production",
    // watch enables webpack's Watch flag, which means it will run endlessly and recompile on saves
    // use webpack --watch instead
    watch: false,

    optimization: {
        runtimeChunk: "single",
        splitChunks: {
            chunks: "all",
            minSize: 0,
            cacheGroups: {
                vendor: {
                    test: /[\\/]node_modules[\\/]/,
                    // Generate a separate bundle file for each vendor.
                    // Returns the name of the bundle file. "npm.[packageName].[contenthash].js"
                    name(module) {
                        // Get the package name from the path.
                        const packageName = module.context.match(/[\\/]node_modules[\\/](.*?)([\\/]|$)/)[1];
                        // Remove @ from the package name.
                        return `npm.${packageName.replace('@', '')}`;
                    },
                }
            }
        }
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
        new HtmlWebpackPlugin({
            // This plugin will generate a HTML file that includes all the webpack bundles.
            // The file will be placed in the dist folder.
            filename: __dirname + "/assets/index.html",
            template: "index.tmpl.html",
            // publicPath is the path the server will serve bundle files from.
            publicPath: "/static/",
        })
    ],

    module: {
        rules: [
            // All files with a '.ts' or '.tsx' extension will be handled by 'ts-loader'.
            { test: /\.tsx?$/, loader: "ts-loader" },

            // All output '.js' files will have any sourcemaps re-processed by 'source-map-loader'.
            { enforce: "pre", test: /\.js$/, loader: "source-map-loader" },
            { test: /\.css$/i, use: ["style-loader", "css-loader"], },
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
