# How to run Selenium tests

## Required packages

All required packages should automatically be downloaded after you run either `npm ci`, or  `npm install`.

Alternatively you can add Selenium by running `npm install selenium-webdriver`.

## Installing browser specific drivers

For Selenium to work, you need to download drivers for the browsers you want to use.

A list of available drivers can be found [here](https://www.selenium.dev/documentation/webdriver/getting_started/install_drivers/).

For our purposes, we require you to add drivers to your `PATH`.

On Linux and macOS, you can verify that your drivers are installed by running `which <driver_name>`.
On Windows you can verify by running `<driver_name>.exe`

## Configuring how the tests run

### Enabling browsers

After having downloaded the drivers and added them to your `PATH`, you need to modify `config.json` to enable the browsers you want to use. This file is located in `dev/src/__tests__/testHelpers/config.json`.

To enable a browser, add `true` to the browser's key.

Our Selenium tests will try to run in all browsers that are enabled.

### Setting the base URL

To specify the base URL used by the tests, add it to the `BASE_URL` key in `config.json`.

## How to run the tests

Selenium tests require that our frontend is reachable at the URL specified in `config.json`.

Selenium tests are run by:

```shell
npm run test:selenium
```

Alternatively, you can run specific tests by running:

```shell
 npm run test:selenium -- -t "test_name"
```

`test_name` will try to match on the `describe` text in your tests.
