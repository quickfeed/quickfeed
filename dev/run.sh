if [ ! -d "public" ]; then
  mkdir public
fi

if [ ! -f "public/index.html" ]; then
    cp index.html public/index.html
fi

if [ ! -f "public/styles.css" ]; then
    cp styles/styles.css public/styles.css
fi

if grep -q "/dev/dist/bundle.js" public/index.html; then
    sed -i 's/\/dev\/dist\/bundle.js/bundle.js/g' public/index.html
fi

if grep -q "/dev/styles/styles.css" public/index.html; then
    sed -i 's/\/dev\/styles\/styles.css/styles.css/g' public/index.html
fi

npx webpack-dev-server --config webpack.config.js --progress --mode development
