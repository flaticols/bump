package tui

import (
	"github.com/charmbracelet/huh"
)

type askConfirmationOpts struct {
	question string
	yesText  string
	noText   string
	tuiCommonProps
}

type AskConfirmationOpt func(*askConfirmationOpts)

func Yes(text string) AskConfirmationOpt {
	return func(o *askConfirmationOpts) {
		o.yesText = text
	}
}

func No(text string) AskConfirmationOpt {
	return func(o *askConfirmationOpts) {
		o.noText = text
	}
}

func AvoidIf(enabled, defaultValue bool) AskConfirmationOpt {
	return func(o *askConfirmationOpts) {
		o.bypassAndRetDefVal = enabled
		o.defaultValue = defaultValue
	}
}

// AskConfirmation ask user to answer yes or not and return result
func AskConfirmation(q string, opts ...AskConfirmationOpt) (confirm bool) {
	o := askConfirmationOpts{
		question: q,
		yesText:  "Yes",
		noText:   "No",
	}

	for _, opt := range opts {
		opt(&o)
	}

	if o.bypassAndRetDefVal {
		return o.defaultValue
	}

	err := huh.NewConfirm().
		Title(o.question).
		Affirmative(o.yesText). //fmt.Sprintf("Yes remove %s!", tag)
		Negative(o.noText).
		Value(&confirm).
		WithTheme(huh.ThemeBase()).Run()
	if err != nil {
		return false
	}
	return confirm
}
