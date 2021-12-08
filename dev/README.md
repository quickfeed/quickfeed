
## Installation
To install dependencies and compile the frontend run:
### `npm install`

### `webpack` or alternatively `npx webpack`

---
### Going to production
When going live this should changed from `test` to `production`
```
    plugins: [
        new webpack.DefinePlugin({
            'process.env.NODE_ENV': JSON.stringify('test')
        }),
        new webpack.DefinePlugin({
            'process.env.ASSET_PATH': JSON.stringify(ASSET_PATH),
        }),
    ],

```
