package handlers

func (m *MessageHandler) OnHelp() error {
	return m.sendMessage(
		`!grohelp: get help!
!gro <name>: adds an item to your grocery list
!groremove <n>: removes item #n from your grocery list
!groremove <n> <m> <o>...: removes item #n, #m, and #o from your grocery list. You can chain as many items as you want.
!grolist: list all the groceries in your grocery list
!groclear: clears your grocery list
!groedit <n> <new name>: updates item #n to a new name/entry
!grodeets <n>: views the full detail of item #n (e.g. who made the entry)

You can also do !grobulk to add your own grocery list. Format:

` + "```" + `
!grobulk
eggs
Soap 1pc
Liquid soap 500 ml
` + "```" + `

These 3 new items will be added to your existing grocery list!

For more information and/or any other questions, you can get in touch with my maintainer through my GitHub repo: https://github.com/verzac/grocer-discord-bot
`,
	)
}
