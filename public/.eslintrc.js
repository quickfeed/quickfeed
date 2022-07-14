module.exports = {
	"parser": "@typescript-eslint/parser",
	"plugins": ["@typescript-eslint"],
	"extends": [
		"eslint:recommended",
		"plugin:react/all"
    ],
	"parserOptions": {
		"ecmaVersion": 6,
		"ecmaFeatures": {
			"jsx": true
		}
	},
	"rules": {
		"react/function-component-definition": [2, {
			"namedComponents": "arrow-function",
			"unnamedComponents": ["function-expression", "arrow-function"]
		}],
		"react/jsx-filename-extension": [2, { "extensions": [".ts", ".tsx"] }],
		"react/jsx-newline": ["off"],
		"no-undef": ["off"],
		"react/jsx-max-depth": [2, { "max": 4 }]

	}
}
