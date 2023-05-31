package interactions

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

func HandleAnimeComplete(event *handler.AutocompleteEvent) error {
	choices := []discord.AutocompleteChoice{
		discord.AutocompleteChoiceString{
			Name:  "test",
			Value: "test",
		},
	}

	return event.Result(
		choices,
	)
}

var AnimeCompleteCommandName = "anime"
