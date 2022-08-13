package models

type TeleBotCommands struct {
	Commands []TeleBotCommand
}

type TeleBotCommand struct {
	Command     string
	Description string
}
