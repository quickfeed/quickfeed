# How to run Selenium tests

## Required packages

All required packages should automatically be downloaded using:

```shell
npm install
```

## Installing browser specific drivers

For Selenium to work, you need to download drivers for the browsers you want to use.
A list of available drivers can be found [here](https://www.selenium.dev/documentation/webdriver/getting_started/install_drivers/).
To use the Selenium's drivers, they must be in your system's `PATH`.

On Linux and macOS, to verify that the drivers are in your `PATH`, run:

```shell
which <driver_name>
```

On Windows you can verify by running

```shell
<driver_name>.exe
```

## Configuring the Selenium tests

You can enable the specific browser drivers by editing `public/src/__tests__/config.json`.
The Selenium tests will try to run in all enabled browsers.

You can also specify a base URL to be used by the tests in the same `config.json` file.
The base URL must match the URL of the web server used for testing.

## Running the Selenium tests

Selenium tests require that the frontend is reachable at the base URL specified in `config.json`.

To start a web server for testing, run:

```shell
make webpack-dev-server
```

This will print messages to the console ending with something like:

```shell
webpack 5.65.0 compiled successfully in 6204 ms
```

Then in another terminal window, run the actual Selenium tests with:

```shell
make selenium
```

Or:

```shell
cd public
npm run test:selenium
```

Alternatively, you can run specific tests by running:

```shell
npm run test:selenium -- -t "test_name"
```

where `test_name` will try to match on the `describe` text in the tests.
