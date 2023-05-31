package commands

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

type SlashCommand struct {
	Data    discord.SlashCommandCreate
	Handler func(event *handler.CommandEvent) error
}

var SlashCommandList = map[string]SlashCommand{
	AnimeCommandData.Name: {
		Data:    AnimeCommandData,
		Handler: HandleAnimeCommand,
	},

	MangaCommandData.Name: {
		Data:    MangaCommandData,
		Handler: HandleMangaCommand,
	},

	AnimeListCommandData.Name: {
		Data:    AnimeListCommandData,
		Handler: HandleAnimeListCommand,
	},

	MangaListCommandData.Name: {
		Data:    MangaListCommandData,
		Handler: HandleMangaListCommand,
	},

	AboutCommandData.Name: {
		Data:    AboutCommandData,
		Handler: HandleAboutCommand,
	},
}

func GetSlashCommandRegisterData() []discord.ApplicationCommandCreate {
	registerCommands := make([]discord.ApplicationCommandCreate, 0, len(SlashCommandList))
	for _, command := range SlashCommandList {
		registerCommands = append(registerCommands, command.Data)
	}
	return registerCommands
}
