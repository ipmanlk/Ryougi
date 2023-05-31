package interactions

import (
	"fmt"
	"ipmanlk/saika/database"
	"ipmanlk/saika/helpers"
	"ipmanlk/saika/structs"
	"strconv"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

func HandleAnimeRatingModal(event *handler.ModalEvent) error {
	event.DeferUpdateMessage()

	// print variables
	idAnilist, _ := strconv.Atoi(event.Variables["idAnilist"])
	origin := event.Variables["origin"]
	prevPage, _ := strconv.Atoi(event.Variables["prevPage"])

	scoreStr := event.Data.Text("score")

	score, err := strconv.Atoi(scoreStr)

	if err != nil {
		_, err = event.CreateFollowupMessage(discord.MessageCreate{
			Embeds: *helpers.GetErrorEmbed("Please provide a valid rating!"),
			Flags:  discord.MessageFlagEphemeral,
		})
		return err
	}

	// check if the rating is between 0 and 10
	if score < 0 || score > 10 {
		_, err = event.CreateFollowupMessage(discord.MessageCreate{
			Embeds: *helpers.GetErrorEmbed("Please provide a rating between 0 and 10!"),
			Flags:  discord.MessageFlagEphemeral,
		})
		return err
	}

	go func() {
		// get media from db
		media, err := database.GetMediaByIDAnilist(idAnilist)

		if err != nil || media == nil {
			fmt.Printf("Error occurred while getting media from db: %v\n", err)

			_, _ = event.CreateFollowupMessage(discord.MessageCreate{
				Embeds: *helpers.GetErrorEmbed("Error occurred while updating your list"),
				Flags:  discord.MessageFlagEphemeral,
			})
			return
		}

		// save in db
		userMedia, err := database.SaveUserMedia(&structs.UserMedia{
			MediaID:   media.ID,
			UserID:    event.User().ID,
			Score:     score,
			MediaType: media.Type,
		})

		if err != nil {
			fmt.Printf("Error occurred while saving user media: %v\n", err)

			_, _ = event.CreateFollowupMessage(discord.MessageCreate{
				Embeds: *helpers.GetErrorEmbed("Error occurred while updating your list"),
				Flags:  discord.MessageFlagEphemeral,
			})

			return
		}

		_, _ = event.CreateFollowupMessage(discord.MessageCreate{
			Embeds: *helpers.GetDefaultEmbed(fmt.Sprintf("You have rated %s with a score of %d.", media.GetTitle(), userMedia.Score)),
			Flags:  discord.MessageFlagEphemeral,
		})

		if origin == "list" {
			msg := helpers.GetListMediaMessage(media, userMedia, prevPage)
			event.UpdateInteractionResponse(msg)
		}
	}()

	return nil
}
