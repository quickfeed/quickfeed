# How to run Selenium tests

# Plugins needed installments

To install Selenium:

```shell
% npm install selenium-webdriver
```

To install webdriver:

Go to: <https://www.selenium.dev/documentation/webdriver/getting_started/install_drivers/>

Install the webdriver for your browser. The documentation lists several ways of using the drivers.

# How to run the tests

1. Boot up QuickFeed locally.
2. Start the geckodriver.
3. Go to the `src/__tests__` folder.
4. Run the specific test:

```shell
% jest <name of the test>
```
