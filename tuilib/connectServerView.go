package tuilib

import (
	"github.com/marcusolsson/tui-go"
	"log"
)

func EnterServerAddr() (string,string) {

	ip := tui.NewEntry()
	ip.SetFocused(true)
	port := tui.NewEntry()

	form := tui.NewGrid(0, 0)
	form.AppendRow(tui.NewLabel("IP server: "),tui.NewLabel("PORT server: "))
	form.AppendRow(ip,port)

	status := tui.NewStatusBar("Ready.")

	login := tui.NewButton("[Connect]")
	login.OnActivated(func(b *tui.Button) {
		status.SetText("Connected.")
	})


	buttons := tui.NewHBox(
		tui.NewSpacer(),
		tui.NewPadder(12, 0, login),

	)

	window := tui.NewVBox(
		tui.NewPadder(12, 0, tui.NewLabel("Connect to server of CHATP2P.")),
		tui.NewPadder(12, 1, form),
		buttons,
	)
	window.SetBorder(true)

	wrapper := tui.NewVBox(
		tui.NewSpacer(),
		window,
		tui.NewSpacer(),
	)
	content := tui.NewHBox(tui.NewSpacer(), wrapper, tui.NewSpacer())

	root := tui.NewVBox(
		content,
		status,
	)

	tui.DefaultFocusChain.Set(ip,port, login)

	ui, err := tui.New(root)
	if err != nil {
		log.Fatal(err)
	}

	ui.SetKeybinding("Enter", func() {
		ui.Quit()

	})

	if err := ui.Run(); err != nil {
		log.Fatal(err)
	}
	return ip.Text(), port.Text()
}