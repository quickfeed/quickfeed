const webpack = require("webpack")
const HtmlWebpackPlugin = require("html-webpack-plugin")
const Dotenv = require('dotenv-webpack');

module.exports = {
    entry: {
        index: {
            import: "./src/index.tsx",
            dependOn: 'overmind'
        },
        overmind: {
            import: "./src/overmind/index.tsx",
            dependOn: 'shared',
        },
        shared: ["./node_modules/@bufbuild/protobuf", "./node_modules/@bufbuild/connect-web", "overmind"],
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
                        const packageName = module.context.match(/[\\/]node_modules[\\/](.*?)([\\/]|$)/)[1]
                        // Remove @ from the package name.
                        return `npm.${packageName.replace('@', '')}`
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
        extensionAlias: {
            '.js': ['.js', '.ts'],
        },
    },


    plugins: [
        new webpack.ProvidePlugin({
            process: 'process/browser.js'
        }),
        new Dotenv(),
        new HtmlWebpackPlugin({
            // This plugin will generate a HTML file that includes all the webpack bundles.
            // The file will be placed in the dist folder.
            filename: __dirname + "/assets/index.html",
            template: "index.tmpl.html",
            // publicPath is the path the server will serve bundle files from.
            publicPath: "/static/",
        })
    ],

    devServer: {
        devMiddleware: {
            index: true,
            writeToDisk: true
        },
        historyApiFallback: true,
        static: [
            {
                directory: __dirname + "/assets",
                publicPath: "/"
            },
            {
                directory: __dirname + "/assets",
                publicPath: "/assets/",
            },
            {
                directory: __dirname + "/dist",
                publicPath: "/static/",
            }
        ]
    },

    module: {
        rules: [
            // All files with a '.ts' or '.tsx' extension will be handled by 'ts-loader'.
            { test: /\.tsx?$/, loader: "ts-loader" },

            // All output '.js' files will have any sourcemaps re-processed by 'source-map-loader'.
            { enforce: "pre", test: /\.js$/, loader: "source-map-loader" },
            {
                test: /\.s[ac]ss$/i, use: [
                    "style-loader",
                    "css-loader",
                    "sass-loader"
                ]
            },
        ]
    },
}
