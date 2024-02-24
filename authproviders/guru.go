package authproviders

import (
	"log"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/defaults"
	"github.com/go-rod/rod/lib/input"
)

// DOM selector which tell that the login process is done and guru page are display
const selectorLogged = "button[data-cy='tab--cloud-sandboxes']"

type AuthGuru struct{}

func (a *AuthGuru) AuthId() string {
	return "guru"
}

func (a *AuthGuru) Login(acgSandboxesUrl string, username string, password string) *rod.Page {
	browser := rod.New().MustConnect().NoDefaultDevice()
	page := browser.MustPage(acgSandboxesUrl)

	if page.MustHas(selectorLogged) {
		return page
	}

	page.MustElement("input[name='email']").MustInput(username).MustType(input.Enter)
	page.MustElement("input[name='password']").MustInput(password).MustType(input.Enter)
	page.MustWaitDOMStable()

	if page.MustHas("input[name='captcha']") {
		if defaults.Show {
			page.MustElement("input[name='captcha']").MustFocus()
		} else {
			log.Fatalf("Warning: CAPTCHA in login form, use -rod=show option")
		}
	}

	if page.MustHas("div.auth0-global-message-error") {
		errorMessage := page.MustElement("div.auth0-global-message-error").MustText()
		log.Fatalf("Error: %s", errorMessage)
	}

	return page
}
