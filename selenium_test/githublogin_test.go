package selenium_test

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tebeka/selenium"
)

// IMPORTANT!
// Test case are made like explained below.
// For example:

// Output:
// <your wanted situation here>
// <your wanted situation here>
// <your wanted situation here>
// Program exited.

//	To run this test:
//	go test -test.run=GithubLogin$

const (
	TAKE_SCREENSHOT = false
	WANT_MESSAGE    = "Hi,"
	USERNAME        = ""
	PASSWORD        = ""
)

// Saves a image taken by the webdriver, and saves it to the current
// folder.
func saveScreenshot(wd selenium.WebDriver, name string) {
	if !TAKE_SCREENSHOT {
		return
	}
	screenshot, err := wd.Screenshot()
	if err != nil {
		logrus.Error("Failed to take screenshot")
	}

	img, _, _ := image.Decode(bytes.NewReader(screenshot))
	out, err := os.Create("./" + name + ".png")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = png.Encode(out, img)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func sleep() {
	time.Sleep(time.Millisecond * 1000)
}

func TestGithubLogin(t *testing.T) {
	// Start a Selenium WebDriver server instance (if one is not already
	// running).
	const (
		seleniumPath    = "../vendor/seleniumhq.org/selenium-server-standalone-3.13.0.jar"
		geckoDriverPath = "../vendor/seleniumhq.org/geckodriver"
		port            = 8080
	)
	opts := []selenium.ServiceOption{
		selenium.StartFrameBuffer(),           // Start an X frame buffer for the browser to run in.
		selenium.GeckoDriver(geckoDriverPath), // Specify the path to GeckoDriver in order to use Firefox.
		selenium.Output(os.Stderr),            // Output debug information to STDERR.
	}
	selenium.SetDebug(true)
	service, err := selenium.NewSeleniumService(seleniumPath, port, opts...)
	if err != nil {
		logrus.Fatal(err)
	}
	defer service.Stop()
	// Connect to the WebDriver instance running locally.
	caps := selenium.Capabilities{"browserName": "firefox", "acceptSslCerts": true}
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
	if err != nil {
		panic(err)
	}
	defer wd.Quit()

	// SETUP DONE
	// Navigate to the index page.
	if err := wd.Get("https://itest.run"); err != nil {
		logrus.WithFields(logrus.Fields{
			"Function": "Get(\"https://www.itest.run\")",
		}).Fatal(err)
	}
	// Need time to load the index screen.
	// Find out why wd.SetPageLoadTimeout(150000) is not working!
	time.Sleep(time.Millisecond * 1000)
	// One way to get image of how the page you are currently at
	// looks like.
	saveScreenshot(wd, "indexpage")

	// Get a reference to the github-login button.
	/*
		elem, err := wd.FindElement(selenium.ByPartialLinkText, "github")
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"Function": "FindElement(loginbox))",
			}).Fatal(err)
		}

		// Click the login button.
		if err := elem.Click(); err != nil {
			logrus.WithFields(logrus.Fields{
				"Function": "Loginbutton Click()",
			}).Fatal(err)
		}*/
	// Quick hack since I didnt find out how to get the above to work
	//wd.ExecuteScript("document.getElementsByClassName(\"social-login\")[0].children[0].firstChild.click()", nil)
	login, err := wd.FindElement(selenium.ByCSSSelector, "a[href=\"/app/login/login/github\"]")
	if err != nil {
		logrus.Fatal(err)
	}
	login.Click()
	sleep()

	// Fetch the fields needed.
	loginField, err := wd.FindElement(selenium.ByID, "login_field")
	if err != nil {
		logrus.Fatal(err)
	}

	loginField.SendKeys(USERNAME)

	passwordField, err := wd.FindElement(selenium.ByID, "password")
	if err != nil {
		logrus.Fatal(err)
	}
	passwordField.SendKeys(PASSWORD)

	loginButton, err := wd.FindElement(selenium.ByName, "commit")
	if err != nil {
		logrus.Fatal(err)
	}

	// One way to get image of how the page you are currently at
	// looks like.
	saveScreenshot(wd, "loginpage")

	loginButton.Click()
	sleep()

	// One way to get image of how the page you are currently at
	// looks like.
	saveScreenshot(wd, "AuthPage")

	authorize, err := wd.FindElement(selenium.ByID, "js-oauth-authorize-btn")
	if err != nil {
		//logrus.Info("No auth page found")
	} else {
		authorize.Click()
	}
	sleep()
	hellomsg, err := wd.FindElement(selenium.ByCSSSelector, "div[class='centerblock container']")
	if err != nil {
		logrus.Fatal(err)
	}
	outputText, err := hellomsg.Text()
	if err != nil {
		logrus.Error("Could not get output text")
	}
	//logrus.Info(outputText)

	fmt.Printf("%s\n", outputText)

	if !reflect.DeepEqual(outputText, WANT_MESSAGE) {
		t.Errorf("have database course %+v want %+v", outputText, WANT_MESSAGE)
	}

}
