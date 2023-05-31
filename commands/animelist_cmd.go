package commands

import (
	"ipmanlk/saika/helpers"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

func HandleAnimeListCommand(event *handler.CommandEvent) error {
	event.DeferCreateMessage(true)

	go func() {
		message := helpers.GetInitialMediaListMessage(event.User(), "ANIME")
		_, _ = event.UpdateInteractionResponse(message)
	}()

	return nil
}

var AnimeListCommandData = discord.SlashCommandCreate{
	Name:        "anime-lists",
	Description: "View your anime lists",
}
