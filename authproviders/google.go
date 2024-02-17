package authproviders

import (
	"log"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/defaults"
	"github.com/go-rod/rod/lib/input"
)

type AuthGoogle struct{}

func (a *AuthGoogle) AuthId() string {
	return "google"
}

func (a *AuthGoogle) Login(acgSandboxesUrl string, username string, password string) *rod.Page {
	browser := rod.New().MustConnect().NoDefaultDevice()
	page := browser.MustPage(acgSandboxesUrl)
	page.MustElement("input[name='email']").MustInput(username).MustType(input.Enter)
	page.MustElement("input[name='password']").MustInput(password).MustType(input.Enter)
	if page.MustHas("input[name='captcha']") {
		if defaults.Show {
			page.MustElement("input[name='captcha']").MustFocus()
		} else {
			log.Fatalf("Warning: CAPTCHA in login form, use -rod=show option")
		}
	}
	return page
}
