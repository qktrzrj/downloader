package main

import "github.com/andlabs/ui"

func completePage() ui.Control {
	vbox := ui.NewVerticalBox()
	vbox.SetPadded(true)
	return vbox
}
