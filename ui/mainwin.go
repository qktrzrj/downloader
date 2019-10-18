package ui

import (
	"gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"golang.org/x/exp/shiny/materialdesign/icons"
	"log"
)

var (
	lineEditor = new(widget.Editor)
	addButton  = new(widget.Button)
	list       = &layout.List{
		Axis: layout.Vertical,
	}
	icon *material.Icon
)

func SetUI() {
	ic, err := material.NewIcon(icons.ContentAdd)
	if err != nil {
		log.Fatal(err)
	}
	icon = ic
	mainwin := app.NewWindow()
	if err := loop(mainwin); err != nil {
		log.Fatal(err)
	}
}

func loop(window *app.Window) error {
	th := material.NewTheme()
	gtx := &layout.Context{
		Queue: window.Queue(),
	}
	for {
		e := <-window.Events()
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			gtx.Reset(e.Config, e.Size)
			download(gtx, th)
			e.Frame(gtx.Ops)
		}
	}
}

func download(gtx *layout.Context, th *material.Theme) {
	gtx.Constraints.Width.Min = 0
	gtx.Constraints.Height.Min = 0
	widgets := []func(){
		func() {
			flex := layout.Flex{
				Spacing:   0,
				Alignment: layout.Middle,
			}
			in := layout.UniformInset(unit.Dp(0))
			button := flex.Rigid(gtx, func() {
				in.Layout(gtx, func() {
					for addButton.Clicked(gtx) {

					}
					th.IconButton(icon).Layout(gtx, addButton)
				})
			})
			input := flex.Rigid(gtx, func() {
				in.Layout(gtx, func() {
					th.Editor("url").Layout(gtx, lineEditor)
				})
			})
			flex.Layout(gtx, button, input)
		},
	}
	list.Layout(gtx, len(widgets), func(i int) {
		layout.UniformInset(unit.Dp(16)).Layout(gtx, widgets[i])
	})
}
