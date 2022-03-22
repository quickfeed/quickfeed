# Testing with Jest

To run the frontend tests in `dev/src/__tests__`, make sure to install the required packages:

```shell
% cd dev
% npm ci
```

To run `jest` from the command line, you will need to install it globally:

```shell
% npm i --global jest
```

For more information on running `jest` from the command line, please see the [getting started](https://jestjs.io/docs/getting-started) documentation.

To run all tests:

```shell
% cd dev/src/__tests__
% jest
```

To run a specific test:

```shell
% jest <test-filename>
```
