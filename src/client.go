package main

import (
	"github.com/marcusolsson/tui-go"
	"log"
	"os"
	"regexp"
	"time"

)
var LoginFinish = make(chan bool)
var InputChat = make(chan string)
var OutPutChat = make(chan string)
var Refresh=make (chan bool)
func main() {
	go LoginThenProcess()
	<-LoginFinish
	ChatFrame:= tui.NewLabel("SIMPLE CHAT\nUser: "+myName+"\n---------------------------------------")
	InputMessage:=tui.NewLabel("Input message\n--------------------------------------- ")
	history := tui.NewVBox()
	historyScroll := tui.NewScrollArea(history)
	historyScroll.SetAutoscrollToBottom(true)
	historyBox := tui.NewVBox(ChatFrame,historyScroll)
	historyBox.SetBorder(true)

	input := tui.NewEntry()
	input.SetFocused(true)
	inputBox := tui.NewVBox(InputMessage,input)
	inputBox.SetBorder(true)

	input.OnSubmit(func(e *tui.Entry) {
		a:=regexp.MustCompile(` `)
		if(e.Text()=="exit"){
			time.Sleep(100*time.Millisecond)
			OutPutChat<-"...exited CHAT P2P..."
			time.Sleep(2*time.Second)
			os.Exit(3)
		}
		contentInForm:=a.Split(e.Text(),2)
		if len(contentInForm)==2 {
			history.Append(tui.NewVBox(
				tui.NewLabel(`                                          TO `+contentInForm[0]+":\t "+contentInForm[1]),
				tui.NewSpacer(),
			))
			InputChat<-e.Text()
			input.SetText("")
		}
	})

	box := tui.NewVBox(historyBox, inputBox)
	ui, err := tui.New(box)
	if err != nil {
		log.Fatal(err)
	}
	go show(history,ui)
	ui.SetKeybinding("Esc", func() { ui.Quit() })
	if err := ui.Run(); err != nil {
		log.Fatal(err)
	}
}
func show(history *tui.Box,ui tui.UI) {
	for {
		showText:=<-OutPutChat
		history.Append(tui.NewVBox(
			tui.NewLabel(showText),
			tui.NewSpacer(),
		))
		time.Sleep(50*time.Millisecond)
		ui.Repaint()
	}
}