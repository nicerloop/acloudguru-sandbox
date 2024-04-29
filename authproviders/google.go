package authproviders

import (
	"fmt"
	"log"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/launcher"
)

type AuthGoogle struct{}

func (a *AuthGoogle) AuthId() string {
	return "google"
}

func (a *AuthGoogle) Login(acgSandboxesUrl string, username string, password string) *rod.Page {
	// Rod user mode is needed to avoid Google bot detection .Set("headless")
	u := launcher.NewUserMode().MustLaunch()
	browser := rod.New().ControlURL(u).MustConnect().NoDefaultDevice()
	//defer browser.MustClose()
	page := browser.MustPage(acgSandboxesUrl).MustWindowMinimize()

	// Wait end of possible autologin
	page.MustWaitDOMStable()
	//time.Sleep(time.Duration(250+rand.Intn(100)) * time.Millisecond)

	// If already logged
	if page.MustHas(selectorLogged) {
		return page
	}

	page.MustElement("a[data-provider='google-oauth2']").MustClick()
	page.MustWaitDOMStable()

	wait := 0
	for {
		wait++

		// End of loop condition
		if page.MustHas(selectorLogged) {
			break
		} else if wait > 600 {
			log.Fatalf("Error: cannot connect")
		}

		// Handle google pages auth
		page = a.HandlePage(acgSandboxesUrl, username, password, page)

		// Error page
		if page.MustHasR("h1[id='headingText']", ".*wrong.*") || page.MustHasR("h1[id='headingText']", ".*error.*") {
			errorTitle := page.MustElement("h1[id='headingText']").MustText()
			errorMessage := page.MustElement("div[id='headingSubtext']").MustText()
			log.Fatalf("Error: %s - %s", errorTitle, errorMessage)
		}

		page.MustWaitDOMStable()
	}

	return page
}

func (a *AuthGoogle) HandlePage(acgSandboxesUrl string, username string, password string, page *rod.Page) *rod.Page {
	// google change element attribute dynamicaly, if happend we ignore and retry
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("error is %v\n", err)
		}
	}()

	// Account selection page
	if page.MustHas(fmt.Sprintf("div[role='link'][data-identifier='%s']", username)) {
		page.MustElement(fmt.Sprintf("div[role='link'][data-identifier='%s']", username)).MustClick()
	}

	// Login page
	if page.MustHas("input[name='identifier']:not([aria-hidden='true'])") {
		page.MustElement("input[name='identifier']:not([aria-hidden='true'])").MustInput(username).MustType(input.Enter)
	}

	// Password page
	if page.MustHas("input[name='Passwd']") {
		page.MustElement("input[name='Passwd']").MustInput(password).MustType(input.Enter)
	}

	// Wait M2F page
	if page.MustHas("iframe") {
		fmt.Printf("\nManual action needed\n => Wainting MFA validation on your configured device (phone, tablet...)\n")
		// Must wait for MFA finish
		page.MustWaitOpen()
	}

	return page
}
