package ui

import "github.com/andlabs/ui"

func settingPage() ui.Control {
	vbox := ui.NewVerticalBox()
	vbox.SetPadded(true)
	return vbox
}
