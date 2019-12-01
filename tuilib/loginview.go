package tuilib

import (
	"github.com/marcusolsson/tui-go"
	"log"
)

func Login() string {

	user := tui.NewEntry()
	user.SetFocused(true)

	form := tui.NewGrid(0, 0)
	form.AppendRow(tui.NewLabel("User"))
	form.AppendRow(user)

	status := tui.NewStatusBar("Ready.")

	login := tui.NewButton("[Login]")
	login.OnActivated(func(b *tui.Button) {
		status.SetText("Logged in.")
	})


	buttons := tui.NewHBox(
		tui.NewSpacer(),
		tui.NewPadder(12, 0, login),

	)

	window := tui.NewVBox(
		tui.NewPadder(12, 0, tui.NewLabel("Welcome to CHAT P2P! Login.")),
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

	tui.DefaultFocusChain.Set(user, login)

	ui, err := tui.New(root)
	if err != nil {
		log.Fatal(err)
	}

	ui.SetKeybinding("Enter", func() {
		ui.Quit()
		return
	})

	if err := ui.Run(); err != nil {
		log.Fatal(err)
	}
	return user.Text()
}