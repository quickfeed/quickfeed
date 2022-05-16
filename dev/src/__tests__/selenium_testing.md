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

## Enabling Browsers

After having downloaded the drivers and added them to your `PATH`, you need to modify `browser.json` to enable the browsers you want to use. This file is located in `dev/src/__tests__/testHelpers/browsers.json`.

To enable a browser, add `true` to the browser's key.

Our Selenium tests will try to run in all browsers that are enabled.

## How to run the tests

>TODO: Fill out this section

Selenium tests require that our frontend is reachable in a browser.

Selenium tests are run by:

```shell
npm run test:selenium
```

Alternatively, you can run specific tests by running:

```shell
 npm run test:selenium -- -t "test_name"
```

`test_name` will try to match on the `describe` text in your tests.
