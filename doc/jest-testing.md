# Testing with Jest

To run the frontend tests in `public/src/__tests__`, make sure to install the required packages:

```shell
% cd public
% npm ci
```

To run all tests using:

```shell
% cd public/src/__tests__
% npm test
```

To run a specific test using:

```shell
% npm test -- <test-filename>
```

`npm test` will run all tests using the `jest` package.

If you wish to run `jest` directly from the command line, you will need to install it globally:

```shell
% npm i --global jest
```

For more information on running `jest` from the command line, please see the [getting started](https://jestjs.io/docs/getting-started) documentation.

To run all tests using `jest`:

```shell
% cd public/src/__tests__
% jest
```

To run a specific test using `jest`:

```shell
% jest <test-filename>
```
