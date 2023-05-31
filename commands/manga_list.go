package commands

import (
	"ipmanlk/saika/helpers"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

func HandleMangaListCommand(event *handler.CommandEvent) error {
	event.DeferCreateMessage(true)

	go func() {
		message := helpers.GetInitialMediaListMessage(event.User(), "MANGA")

		_, _ = event.UpdateInteractionResponse(message)
	}()

	return nil
}

var MangaListCommandData = discord.SlashCommandCreate{
	Name:        "manga-lists",
	Description: "View your manga lists",
}
