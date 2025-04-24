/** @type {import('tailwindcss').Config} */
module.exports = {
    content: [
        './src/**/*.{js,jsx,ts,tsx}',
        './index.tmpl.html',
    ],
    plugins: [require('daisyui')],
}
