## WebPack
### Keeping it close to the QF Repo
So I have moved every css file into a single one, outside the WebPack entry point.

If you want to use css files inside of `/src` you need to add webpack's `css-loader` and edit the config file.

Otherwise it won't compile.

---

### Purpose of this branch
This is just a simple branch to maintain the structure of original QF repo, as well as serving as a simple way to get webpack to work. 

*Might add more here*

---

### Going to production
When going live this should be from `test` to `production`
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

---

## Process for running from npm build
### `npm install`

### `webpack`



