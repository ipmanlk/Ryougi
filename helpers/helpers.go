package helpers

import (
	"fmt"
	"ipmanlk/saika/database"
	"ipmanlk/saika/structs"
	"log"
	"math"
	"strings"

	"github.com/disgoorg/disgo/discord"
)

func GetErrorEmbed(message string) *[]discord.Embed {
	embed := discord.NewEmbedBuilder()
	embed.SetColor(0xFF4500)
	embed.SetDescription(message)

	return &[]discord.Embed{
		embed.Build(),
	}
}

func GetDefaultEmbed(message string) *[]discord.Embed {
	embed := discord.NewEmbedBuilder()
	embed.SetColor(0x7B1FA2)
	embed.SetDescription(message)

	return &[]discord.Embed{
		embed.Build(),
	}
}

func GetMediaSearchMessage(
	results *[]structs.AnilistMedia,
	page int,
	searchQueryID string,
	userID string,
	mediaType string,
) discord.MessageUpdate {

	media := (*results)[page-1]
	pages := len(*results)
	isAnime := mediaType == "ANIME"

	msg := discord.NewMessageUpdateBuilder()

	embed := discord.NewEmbedBuilder()
	embed.SetColor(0xFF4081)
	embed.SetTitle(media.GetTitle())
	embed.SetDescription(media.GetDescription())
	embed.SetURL(media.SiteUrl)
	embed.SetThumbnail(media.CoverImage.Large)
	embed.SetImage(media.BannerImage)
	embed.AddField("Score", media.GetMeanScoreStr(), true)
	embed.AddField("Status", media.GetStatus(), true)

	if isAnime {
		embed.AddField("Episodes", media.GetEpisodesStr(), true)
	} else {
		embed.AddField("Chapters", media.GetChapterStr(), true)
		embed.AddField("Volumes", media.GetVolumeStr(), true)
	}

	embed.AddField("Year", media.GetSeasonYearStr(), true)
	embed.AddField("Format", media.GetFormat(), true)
	embed.AddField("Source", media.GetSource(), true)
	embed.AddField("Genres", media.GetGenres(), false)
	embed.AddField("Tags", media.GetTags(), false)

	if len(*results) > 2 {
		embed.SetFooterText(fmt.Sprintf("Page %d of %d", page, pages))
	}

	msg.AddEmbeds(embed.Build())

	selectMenuPlaceholder := "Add to anime lists"

	if !isAnime {
		selectMenuPlaceholder = "Add to manga lists"
	}

	selectMenuOptions := []discord.StringSelectMenuOption{}
	selectMenuOptions = append(selectMenuOptions, discord.NewStringSelectMenuOption("Add to Planning", fmt.Sprintf("sm_media_action/%d/PLANNING", media.IdAnilist)))

	if isAnime {
		selectMenuOptions = append(selectMenuOptions, discord.NewStringSelectMenuOption("Add to Watching", fmt.Sprintf("sm_media_action/%d/CURRENT", media.IdAnilist)))
	} else {
		selectMenuOptions = append(selectMenuOptions, discord.NewStringSelectMenuOption("Add to Reading", fmt.Sprintf("sm_media_action/%d/CURRENT", media.IdAnilist)))
	}

	selectMenuOptions = append(selectMenuOptions, discord.NewStringSelectMenuOption("Add to Completed", fmt.Sprintf("sm_media_action/%d/COMPLETED", media.IdAnilist)))
	selectMenuOptions = append(selectMenuOptions, discord.NewStringSelectMenuOption("Add to Paused", fmt.Sprintf("sm_media_action/%d/PAUSED", media.IdAnilist)))
	selectMenuOptions = append(selectMenuOptions, discord.NewStringSelectMenuOption("Add to Dropped", fmt.Sprintf("sm_media_action/%d/DROPPED", media.IdAnilist)))

	if isAnime {
		selectMenuOptions = append(selectMenuOptions, discord.NewStringSelectMenuOption("Add to Repeating", fmt.Sprintf("sm_media_action/%d/REPEATING", media.IdAnilist)))
	} else {
		selectMenuOptions = append(selectMenuOptions, discord.NewStringSelectMenuOption("Add to Rereading", fmt.Sprintf("sm_media_action/%d/REPEATING", media.IdAnilist)))
	}

	selectMenu := discord.NewStringSelectMenu("sm_media_action", selectMenuPlaceholder,
		selectMenuOptions...,
	)

	msg.AddActionRow(
		selectMenu,
	)

	if pages == 1 {
		return msg.Build()
	}

	navigationComponents := []discord.InteractiveComponent{}

	if page > 1 {
		prevButton := discord.NewPrimaryButton("Prev", fmt.Sprintf("btn_media_results/%s/%s/pages/%d", userID, searchQueryID, page-1))
		navigationComponents = append(navigationComponents, prevButton)
	}

	if page < pages {
		prevButton := discord.NewPrimaryButton("Next", fmt.Sprintf("btn_media_results/%s/%s/pages/%d", userID, searchQueryID, page+1))
		navigationComponents = append(navigationComponents, prevButton)
	}

	msg.AddActionRow(navigationComponents...)

	return msg.Build()
}

func GetInitialMediaListMessage(user discord.User, mediaType string) discord.MessageUpdate {
	userMedia, err := database.GetAllUserMedia(user.ID, mediaType, "")
	if err != nil {
		log.Printf("Error getting user media: %s", err.Error())
		return discord.MessageUpdate{
			Embeds: GetErrorEmbed("Error getting user media"),
		}
	}

	if userMedia == nil {
		userMedia = []structs.UserMedia{}
	}

	mediaCounts := map[string]int{"PLANNING": 0, "CURRENT": 0, "COMPLETED": 0, "PAUSED": 0, "DROPPED": 0, "REPEATING": 0}
	for _, media := range userMedia {
		mediaCounts[media.Status]++
	}

	isAnime := mediaType == "ANIME"
	currentStatus, repeatingStatus := "Watching", "Repeating"
	if !isAnime {
		currentStatus, repeatingStatus = "Reading", "Rereading"
	}

	readableMediaType := strings.Title(strings.ToLower(mediaType))

	embed := discord.NewEmbedBuilder().
		SetTitle(fmt.Sprintf("%s's %s List", user.Username, readableMediaType)).
		SetDescription(fmt.Sprintf("You have %d %s in your lists", len(userMedia), readableMediaType)).
		SetColor(0xFF4081).
		SetImage("https://i.imgur.com/GArVV8M.jpg").
		AddField("Planning", fmt.Sprintf("%d", mediaCounts["PLANNING"]), true).
		AddField(currentStatus, fmt.Sprintf("%d", mediaCounts["CURRENT"]), true).
		AddField("Completed", fmt.Sprintf("%d", mediaCounts["COMPLETED"]), true).
		AddField("Paused", fmt.Sprintf("%d", mediaCounts["PAUSED"]), true).
		AddField("Dropped", fmt.Sprintf("%d", mediaCounts["DROPPED"]), true).
		AddField(repeatingStatus, fmt.Sprintf("%d", mediaCounts["REPEATING"]), true)

	selectMenuPlaceholder := fmt.Sprintf("View %s lists", strings.ToLower(mediaType))
	selectMenuOptions := []discord.StringSelectMenuOption{}

	for status, count := range mediaCounts {
		if count == 0 {
			continue
		}

		menuOptionLabel := fmt.Sprintf("View %s", strings.Title(strings.ToLower(status)))

		if status == "CURRENT" {
			menuOptionLabel = fmt.Sprintf("View %s", strings.Title(strings.ToLower(currentStatus)))
		}

		if status == "REPEATING" {
			menuOptionLabel = fmt.Sprintf("View %s", strings.Title(strings.ToLower(repeatingStatus)))
		}

		selectMenuOptions = append(selectMenuOptions, discord.NewStringSelectMenuOption(menuOptionLabel, fmt.Sprintf("sm_media_lists/%s/%s", mediaType, status)))
	}

	mb := discord.NewMessageUpdateBuilder().
		SetEmbeds(embed.Build())

	if len(selectMenuOptions) > 0 {
		mb.AddActionRow(discord.NewStringSelectMenu("sm_media_lists", selectMenuPlaceholder, selectMenuOptions...))
	}

	return mb.Build()
}

func GetMediaListMessage(user discord.User, userMedia *[]structs.UserMedia, page int) discord.MessageUpdate {
	userMediaSlice := *userMedia
	userMediaType := userMediaSlice[0].MediaType
	userMediaStatus := userMediaSlice[0].Status
	userMediaFormattedStatus := userMediaSlice[0].GetStatus()
	pages := int(math.Ceil(float64(len(userMediaSlice)) / 10))

	embed := discord.NewEmbedBuilder().
		SetTitle(fmt.Sprintf("%s's %s %s List", user.Username, userMediaFormattedStatus, userMediaSlice[0].GetType())).
		SetColor(0xFF4081).
		SetFooterText(fmt.Sprintf("Page %d of %d", page, pages))

	selectMenuPlaceholder := map[bool]string{true: "Select anime to view", false: "Select manga to view"}[userMediaType == "ANIME"]
	selectMenuOptions := []discord.StringSelectMenuOption{}
	var descEntries []string

	for i := (page - 1) * 10; i < page*10 && i < len(userMediaSlice); i++ {
		u := userMediaSlice[i]
		media, err := database.GetMediaByObjectID(u.MediaID)
		if err != nil {
			log.Printf("Error getting media by ID: %s", err.Error())
			continue
		}
		scoreText := map[bool]string{true: fmt.Sprintf(" - **%d/10**", u.Score), false: ""}[u.Score > 0]
		descEntries = append(descEntries, fmt.Sprintf("%d. %s%s", i+1, media.GetTitle(), scoreText))

		selectMenuMediaTitle := func(title string) string {
			if len(title) > 80 {
				return title[:80] + "..."
			}
			return title
		}(media.GetTitle())

		selectMenuOptions = append(selectMenuOptions, discord.NewStringSelectMenuOption(
			fmt.Sprintf("%d. %s", i+1, selectMenuMediaTitle),
			fmt.Sprintf("%s/%d", u.ID.Hex(), page),
		))
	}

	msg := discord.NewMessageUpdateBuilder().
		SetEmbeds(embed.SetDescription(strings.Join(descEntries, "\n")).Build()).
		AddActionRow(discord.NewStringSelectMenu("sm_media_list", selectMenuPlaceholder, selectMenuOptions...))

	if navigationComponents := func() []discord.InteractiveComponent {
		nc := []discord.InteractiveComponent{}
		if page > 1 {
			nc = append(nc, discord.NewPrimaryButton("Prev", fmt.Sprintf("btn_media_list/%s/%s/%d", userMediaType, userMediaStatus, page-1)))
		}
		if page < pages {
			nc = append(nc, discord.NewPrimaryButton("Next", fmt.Sprintf("btn_media_list/%s/%s/%d", userMediaType, userMediaStatus, page+1)))
		}
		return nc
	}(); len(navigationComponents) > 0 {
		msg.AddActionRow(navigationComponents...)
	}

	// add back to list button
	msg.AddActionRow(discord.NewSecondaryButton("Back to lists", fmt.Sprintf("btn_media_lists/%s", userMediaType)))

	return msg.Build()
}

func GetListMediaMessage(media *structs.AnilistMedia, userMedia *structs.UserMedia, prevPage int) discord.MessageUpdate {
	userMediaType := userMedia.MediaType
	userMediaStatus := userMedia.Status

	isAnime := userMediaType == "ANIME"

	embed := discord.NewEmbedBuilder().
		SetTitle(media.GetTitle()).
		SetColor(0xFF4081).
		SetImage(media.BannerImage)

	if isAnime {
		embed.AddField("Duration", fmt.Sprintf("%d minutes", media.Duration), true)
		embed.AddField("Episodes", fmt.Sprintf("%d", media.Episodes), true)
	}

	if !isAnime {
		embed.AddField("Chapters", fmt.Sprintf("%d", media.Chapters), true)
		embed.AddField("Volumes", fmt.Sprintf("%d", media.Volumes), true)
	}

	embed.AddField("Status", media.GetStatus(), true)

	embed.AddField("Your Status", userMedia.GetStatus(), true)

	if userMedia.Score > 0 {
		embed.AddField("Your Score", fmt.Sprintf("%d/10", userMedia.Score), true)
	}

	rateButton := discord.NewSuccessButton("Rate", fmt.Sprintf("btn_rate/list/%s/%d/%d", media.Type, media.IdAnilist, prevPage))
	deleteButton := discord.NewDangerButton("Remove from list", fmt.Sprintf("btn_delete/%s/%s/%s/%d", userMedia.Status, userMedia.ID.Hex(), media.Type, prevPage))

	backButton := discord.NewSecondaryButton("Back to list", fmt.Sprintf("btn_media_list/%s/%s/%d", userMediaType, userMediaStatus, prevPage))

	msg := discord.NewMessageUpdateBuilder().
		SetEmbeds(embed.Build()).
		AddActionRow(rateButton, deleteButton).
		AddActionRow(backButton).
		Build()

	return msg
}
