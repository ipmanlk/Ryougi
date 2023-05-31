package commands

import (
	"fmt"
	"runtime"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

func HandleAboutCommand(event *handler.CommandEvent) error {
	event.DeferCreateMessage(true)

	embed := discord.NewEmbedBuilder()
	embed.SetColor(0xFF4500)
	embed.SetTitle("Ryougi")
	embed.SetDescription("A terrible premonition can bring about a terrible reality.")
	embed.AddField("Language", fmt.Sprintf("%s", runtime.Version()), true)
	embed.AddField("Library", fmt.Sprintf("Disgo [%s]", disgo.Version), true)
	embed.AddField("Guilds", fmt.Sprintf("%d", event.Client().Caches().GuildsLen()), false)
	embed.SetImage("https://media.tenor.com/2rI7gwSzYYEAAAAd/idleglance-amv.gif")
	embeds := []discord.Embed{embed.Build()}

	event.UpdateInteractionResponse(discord.MessageUpdate{
		Embeds: &embeds,
	})

	return nil
}

var AboutCommandData = discord.SlashCommandCreate{
	Name:        "about",
	Description: "View information about the bot",
}
