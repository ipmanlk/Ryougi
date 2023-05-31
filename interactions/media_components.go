package interactions

import (
	"fmt"
	"ipmanlk/saika/database"
	"ipmanlk/saika/helpers"
	"ipmanlk/saika/structs"
	"log"
	"strconv"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

func HandleMediaResultPagination(event *handler.ComponentEvent) error {
	event.DeferUpdateMessage()

	ownerID := event.Variables["ownerID"]

	if ownerID != event.User().ID.String() {
		_, err := event.CreateFollowupMessage(discord.MessageCreate{
			Embeds: *helpers.GetErrorEmbed("You are not allowed to do that"),
			Flags:  discord.MessageFlagEphemeral,
		})
		return err
	}

	// log event variables
	searchQueryId := event.Variables["searchQueryID"]

	// get page and parse it as an int
	page, _ := strconv.Atoi(event.Variables["page"])
	searchQuery, err := database.GetSearchQueryByHexID(searchQueryId)

	if err != nil || searchQuery == nil {
		_, err := event.CreateFollowupMessage(discord.MessageCreate{
			Embeds: *helpers.GetErrorEmbed("Unable to find search query"),
			Flags:  discord.MessageFlagEphemeral,
		})
		return err
	}

	cachedChannel, _ := event.MessageChannel()
	isNsfwChannel := cachedChannel.NSFW()

	// get search results
	results, err := database.SearchMedia(searchQuery.SearchText, searchQuery.MediaType, isNsfwChannel)

	// print results length
	if err != nil {
		_, err := event.CreateFollowupMessage(discord.MessageCreate{
			Embeds: *helpers.GetErrorEmbed("Failed to get search results"),
			Flags:  discord.MessageFlagEphemeral,
		})
		return err
	}

	// if there are no results
	if len(results) == 0 {
		_, err := event.CreateFollowupMessage(discord.MessageCreate{
			Embeds: *helpers.GetErrorEmbed("No results found"),
			Flags:  discord.MessageFlagEphemeral,
		})
		return err
	}

	// update the message with the new embed
	mediaMsg := helpers.GetMediaSearchMessage(&results, page, searchQueryId, event.User().ID.String(), searchQuery.MediaType)
	_, err = event.UpdateInteractionResponse(mediaMsg)

	return err
}

func HandleMediaResultSelectMenu(event *handler.ComponentEvent) error {
	event.DeferUpdateMessage()

	// takes the shape: action_anime/269/completed
	selectedOption := event.StringSelectMenuInteractionData().Values[0]

	// get idAnilist and status from the selectedOption string
	parts := strings.Split(selectedOption, "/")
	idAnilistStr := parts[1]
	status := parts[2]
	idAnilist, _ := strconv.Atoi(idAnilistStr)

	go func() {
		// get media from db
		media, err := database.GetMediaByIDAnilist(idAnilist)

		if err != nil || media == nil {
			log.Printf("Error occurred while getting media from db: %v\n", err)

			_, _ = event.CreateFollowupMessage(discord.MessageCreate{
				Embeds: *helpers.GetErrorEmbed("Error occurred while updating your list"),
				Flags:  discord.MessageFlagEphemeral,
			})
			return
		}

		// save in db
		var userMedia *structs.UserMedia

		userMedia, err = database.SaveUserMedia(&structs.UserMedia{
			MediaID:   media.ID,
			UserID:    event.User().ID,
			Status:    status,
			Score:     -1,
			MediaType: media.Type,
		})

		if err != nil {
			log.Printf("Error occurred while saving user anime: %v\n", err)

			_, _ = event.CreateFollowupMessage(discord.MessageCreate{
				Embeds: *helpers.GetErrorEmbed("Error occurred while updating your list"),
				Flags:  discord.MessageFlagEphemeral,
			})
			return
		}

		if userMedia == nil {
			_, _ = event.CreateFollowupMessage(discord.MessageCreate{
				Embeds: *helpers.GetErrorEmbed("Error occurred while updating your list"),
				Flags:  discord.MessageFlagEphemeral,
			})
			return
		}

		_, _ = event.UpdateInteractionResponse(discord.MessageUpdate{
			Components: &event.Message.Components,
		})

		_, _ = event.CreateFollowupMessage(discord.MessageCreate{
			Embeds: *helpers.GetDefaultEmbed(fmt.Sprintf("%s has been added to your %s list.", media.GetTitle(), strings.ToLower(userMedia.GetStatus()))),
			Flags:  discord.MessageFlagEphemeral,
		})

	}()

	return nil
}

func HandleMediaRateButton(event *handler.ComponentEvent) error {
	mediaType := event.Variables["mediaType"]
	idAnilistStr := event.Variables["idAnilist"]
	origin := event.Variables["origin"]
	prevPage := event.Variables["prevPage"]

	// print variables
	modelTitle := "Rate Anime"

	if mediaType == "MANGA" {
		modelTitle = "Rate Manga"
	}

	model := discord.NewModalCreateBuilder()
	model.CustomID = fmt.Sprintf("model_rate/%s/%s/%s", origin, idAnilistStr, prevPage)
	model.Title = modelTitle
	field := discord.NewTextInput("score", discord.TextInputStyleShort, "Your Score")
	field.Placeholder = "Enter a number between 0 and 10"
	model.AddActionRow(field)

	return event.CreateModal(model.Build())
}

func HandleMediaListsSelectMenu(event *handler.ComponentEvent) error {
	event.DeferUpdateMessage()

	selectedOption := event.StringSelectMenuInteractionData().Values[0]

	parts := strings.Split(selectedOption, "/")
	mediaType := parts[1]
	status := parts[2]

	// get media from db
	media, err := database.GetAllUserMedia(event.User().ID, mediaType, status)

	if err != nil || media == nil {
		log.Printf("Error occurred while getting media from db: %v\n", err)

		_, _ = event.CreateFollowupMessage(discord.MessageCreate{
			Embeds: *helpers.GetErrorEmbed("Error occurred while getting your list"),
			Flags:  discord.MessageFlagEphemeral,
		})
		return err
	}

	message := helpers.GetMediaListMessage(event.User(), &media, 1)

	_, err = event.UpdateInteractionResponse(message)

	return err
}

func HandleMediaListButton(event *handler.ComponentEvent) error {
	event.DeferUpdateMessage()
	message := helpers.GetInitialMediaListMessage(event.User(), event.Variables["mediaType"])
	_, err := event.UpdateInteractionResponse(message)
	return err
}

func HandleMediaListPagination(event *handler.ComponentEvent) error {
	event.DeferUpdateMessage()

	// get page and parse it as an int
	page, _ := strconv.Atoi(event.Variables["page"])

	// get media from db
	media, err := database.GetAllUserMedia(event.User().ID, event.Variables["mediaType"], event.Variables["status"])

	if err != nil || media == nil {
		log.Printf("Error occurred while getting media from db: %v\n", err)

		_, _ = event.CreateFollowupMessage(discord.MessageCreate{
			Embeds: *helpers.GetErrorEmbed("Error occurred while getting your list"),
			Flags:  discord.MessageFlagEphemeral,
		})
		return err
	}

	// update the message with the new embed
	_, err = event.UpdateInteractionResponse(helpers.GetMediaListMessage(event.User(), &media, page))

	return err
}

func HandleMediaListSelectMenu(event *handler.ComponentEvent) error {
	event.DeferUpdateMessage()

	selectedOption := event.StringSelectMenuInteractionData().Values[0]

	parts := strings.Split(selectedOption, "/")
	userMediaID := parts[0]
	prevPage, _ := strconv.Atoi(parts[1])

	// get user media
	userMedia, err := database.GetUserMediaByHexID(userMediaID)

	if err != nil || userMedia == nil {
		log.Printf("Error occurred while getting user media from db: %v\n", err)

		_, _ = event.CreateFollowupMessage(discord.MessageCreate{
			Embeds: *helpers.GetErrorEmbed("Internal error occurred."),
			Flags:  discord.MessageFlagEphemeral,
		})
		return err
	}

	// get media from db
	media, err := database.GetMediaByHexID(userMedia.MediaID.Hex())

	if err != nil || media == nil {
		log.Printf("Error occurred while getting media from db: %v\n", err)

		_, _ = event.CreateFollowupMessage(discord.MessageCreate{
			Embeds: *helpers.GetErrorEmbed("Internal error occurred."),
			Flags:  discord.MessageFlagEphemeral,
		})
		return err
	}

	message := helpers.GetListMediaMessage(media, userMedia, prevPage)
	_, err = event.UpdateInteractionResponse(message)

	return err
}

func HandleMediaListDeleteButton(event *handler.ComponentEvent) error {
	event.DeferUpdateMessage()

	userMediaID := event.Variables["userMediaID"]
	mediaType := event.Variables["mediaType"]
	status := event.Variables["status"]
	prevPage, _ := strconv.Atoi(event.Variables["prevPage"])

	err := database.DeleteUserMediaByHexID(userMediaID)

	if err != nil {
		log.Printf("Error occurred while deleting user media from db: %v\n", err)

		_, _ = event.CreateFollowupMessage(discord.MessageCreate{
			Embeds: *helpers.GetErrorEmbed("Internal error occurred."),
			Flags:  discord.MessageFlagEphemeral,
		})
		return err
	}

	_, _ = event.CreateFollowupMessage(discord.MessageCreate{
		Embeds: *helpers.GetDefaultEmbed("Entry has been deleted from your list."),
		Flags:  discord.MessageFlagEphemeral,
	})

	userMedia, err := database.GetAllUserMedia(event.User().ID, mediaType, status)

	if err != nil {
		log.Printf("Error occurred while getting user media from db: %v\n", err)

		_, _ = event.CreateFollowupMessage(discord.MessageCreate{
			Embeds: *helpers.GetErrorEmbed("Internal error occurred."),
			Flags:  discord.MessageFlagEphemeral,
		})
		return err
	}

	if userMedia != nil && len(userMedia) > 0 {
		msg := helpers.GetMediaListMessage(event.User(), &userMedia, prevPage)
		_, err = event.UpdateInteractionResponse(msg)

		return err
	}

	msg := helpers.GetInitialMediaListMessage(event.User(), mediaType)
	_, err = event.UpdateInteractionResponse(msg)

	return err
}
