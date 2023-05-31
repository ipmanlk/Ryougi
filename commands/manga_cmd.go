package commands

import (
	"fmt"
	"ipmanlk/saika/anilist"
	"ipmanlk/saika/database"
	"ipmanlk/saika/helpers"
	"ipmanlk/saika/structs"
	"log"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

func HandleMangaCommand(event *handler.CommandEvent) error {
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

	log.Println("NSFW Channel: ", isNsfwChannel)

	go func() {
		_, err := anilist.SearchMedia(searchQuery, "MANGA")

		if err != nil {
			fmt.Printf("Error while searching for manga from api: %v", err)
		}

		results, err := database.SearchMedia(searchQuery, "MANGA", isNsfwChannel)

		if err != nil {
			fmt.Printf("Error while searching for manga from db: %v", err)

			_, _ = event.UpdateInteractionResponse(discord.MessageUpdate{
				Embeds: helpers.GetDefaultEmbed("Error occurred while searching for manga"),
			})
			return
		}

		if len(results) == 0 {
			_, _ = event.UpdateInteractionResponse(discord.MessageUpdate{
				Embeds: helpers.GetDefaultEmbed("No results found"),
			})
			return
		}

		// if there are more than one results, store search query
		searchQuery, err := database.SaveSearchQuery(&structs.AnilistSearchQuery{
			SearchText: searchQuery,
			MediaType:  "MANGA",
		})

		if err != nil {
			fmt.Printf("Error while saving search query: %v", err)

			_, _ = event.UpdateInteractionResponse(discord.MessageUpdate{
				Embeds: helpers.GetDefaultEmbed("Error occurred while searching for manga"),
			})
			return
		}

		animeMsg := helpers.GetMediaSearchMessage(&results, 1, searchQuery.ID.Hex(), event.User().ID.String(), "MANGA")
		_, err = event.UpdateInteractionResponse(animeMsg)
	}()

	return nil
}

var MangaCommandData = discord.SlashCommandCreate{
	Name:        "manga",
	Description: "Search for Manga",
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionString{
			Name:        "name",
			Description: "Manga name",
			Required:    true,
		},
	},
}
