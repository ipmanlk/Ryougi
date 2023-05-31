package commands

import (
	"fmt"
	"ipmanlk/saika/anilist"
	"ipmanlk/saika/database"
	"ipmanlk/saika/helpers"
	"ipmanlk/saika/structs"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

func HandleAnimeCommand(event *handler.CommandEvent) error {
	searchQuery := strings.ToLower(event.SlashCommandInteractionData().String("name"))

	event.DeferCreateMessage(false)

	if searchQuery == "" {
		_, error := event.UpdateInteractionResponse(discord.MessageUpdate{
			Embeds: helpers.GetErrorEmbed("Please provide a name!"),
		})
		return error
	}

	cachedChannel, _ := event.MessageChannel()
	isNsfwChannel := cachedChannel.NSFW()

	go func() {
		// check if there are any anime with the given name
		_, err := anilist.SearchMedia(searchQuery, "ANIME")

		if err != nil {
			fmt.Printf("Error while searching for anime from api: %v", err)
		}

		results, err := database.SearchMedia(searchQuery, "ANIME", isNsfwChannel)

		if err != nil {
			fmt.Printf("Error while searching for anime from db: %v", err)

			_, _ = event.UpdateInteractionResponse(discord.MessageUpdate{
				Embeds: helpers.GetDefaultEmbed("Error occurred while searching for anime"),
			})
			return
		}

		if len(results) == 0 {
			_, _ = event.UpdateInteractionResponse(discord.MessageUpdate{
				Embeds: helpers.GetDefaultEmbed(fmt.Sprintf("No results found for \"%s\"", searchQuery)),
			})
			return
		}

		// if there are more than one results, store search query
		searchQuery, err := database.SaveSearchQuery(&structs.AnilistSearchQuery{
			SearchText: searchQuery,
			MediaType:  "ANIME",
		})

		if err != nil {
			fmt.Printf("Error while saving search query: %v", err)

			_, _ = event.UpdateInteractionResponse(discord.MessageUpdate{
				Embeds: helpers.GetDefaultEmbed("Error occurred while searching for anime"),
			})
			return
		}

		animeMsg := helpers.GetMediaSearchMessage(&results, 1, searchQuery.ID.Hex(), event.User().ID.String(), "ANIME")
		_, err = event.UpdateInteractionResponse(animeMsg)
	}()

	return nil
}

var AnimeCommandData = discord.SlashCommandCreate{
	Name:        "anime",
	Description: "Search for Anime",
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionString{
			Name:        "name",
			Description: "Anime name",
			Required:    true,
		},
	},
}
