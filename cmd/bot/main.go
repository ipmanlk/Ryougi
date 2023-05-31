package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"ipmanlk/saika/commands"
	"ipmanlk/saika/config"
	"ipmanlk/saika/database"
	"ipmanlk/saika/interactions"
	"path"

	"os"
	"os/signal"
	"syscall"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgo/handler/middleware"
	"github.com/disgoorg/log"
	"github.com/disgoorg/snowflake/v2"
)

func main() {
	log.SetLevel(log.LevelInfo)
	log.Infof("Bot: [disgo version: %s] Starting....", disgo.Version)

	// Initialize MongoDB
	database.InitMongoDb()

	token := config.GetEnv("BOT_TOKEN", "")
	production := config.GetEnv("PRODUCTION", "0") == "1"

	r := handler.New()
	r.Use(middleware.Logger)

	r.Group(func(r handler.Router) {
		r.Use(middleware.Print("slash commands"))

		r.Command("/anime", commands.HandleAnimeCommand)
		r.Command("/manga", commands.HandleMangaCommand)
		r.Command("/anime-lists", commands.HandleAnimeListCommand)
		r.Command("/manga-lists", commands.HandleMangaListCommand)
		r.Command("/about", commands.HandleAboutCommand)
	})

	r.Group(func(r handler.Router) {
		r.Use(middleware.Print("components"))

		r.Component("btn_media_results/{ownerID}/{searchQueryID}/pages/{page}", interactions.HandleMediaResultPagination)
		r.Component("btn_rate/{origin}/{mediaType}/{idAnilist}/{prevPage}", interactions.HandleMediaRateButton)
		r.Component("btn_delete/{status}/{userMediaID}/{mediaType}/{prevPage}", interactions.HandleMediaListDeleteButton)

		r.Component("sm_media_action", interactions.HandleMediaResultSelectMenu)

		// select menu for selecting a media list
		r.Component("sm_media_lists", interactions.HandleMediaListsSelectMenu)
		// back to lists button
		r.Component("btn_media_lists/{mediaType}", interactions.HandleMediaListButton)
		// pagination for a media list (ex, planning list)
		r.Component("btn_media_list/{mediaType}/{status}/{page}", interactions.HandleMediaListPagination)
		// media select menu from a media list (anime/manga)
		r.Component("sm_media_list", interactions.HandleMediaListSelectMenu)

		r.Modal("model_rate/{origin}/{idAnilist}/{prevPage}", interactions.HandleAnimeRatingModal)
	})

	client, err := disgo.New(token,
		// set gateway options
		bot.WithGatewayConfigOpts(
			gateway.WithIntents(gateway.IntentGuilds),
		),
		bot.WithCacheConfigOpts(
			cache.WithCaches(cache.FlagChannels, cache.FlagGuilds),
		),
		// add event listeners
		bot.WithEventListeners(r, &events.ListenerAdapter{
			OnReady: func(event *events.Ready) {
				// log bot started and mode (production or not)
				log.Infof("Bot started successfully! [PRODUCTION: %v]", production)
			},
		}),
	)

	if err != nil {
		log.Fatal("error while starting bot: ", err)
	}

	updateCommands(client, production)

	defer client.Close(context.TODO())

	// connect to the gateway
	if err = client.OpenGateway(context.TODO()); err != nil {
		log.Fatal("error while connecting to gateway: ", err)
	}

	log.Info("Bot is now running. Press CTRL-C to exit.")

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)
	<-s
}

func updateCommands(botClient bot.Client, production bool) {
	registerData := commands.GetSlashCommandRegisterData()

	// check if commands changed
	commandsChanged := checkCommandsChanged(registerData, production)

	if commandsChanged {
		if production {
			botClient.Rest().SetGlobalCommands(botClient.ApplicationID(), registerData)
			log.Infof("Commands updated in production")
		} else {
			botClient.Rest().SetGuildCommands(botClient.ApplicationID(), snowflake.MustParse(config.GetEnv("GUILD_ID", "")), registerData)
			log.Infof("Commands updated in development")
		}
	} else {
		log.Infof("Commands not updated")
	}
}

func checkCommandsChanged(registerData []discord.ApplicationCommandCreate, production bool) bool {
	bytes, _ := json.Marshal(registerData)
	hash := sha256.Sum256(bytes)
	currentHashStr := hex.EncodeToString(hash[:])

	hashFileName := "cmd-dev"
	if production {
		hashFileName = "cmd-prod"
	}

	cacheFilePath := path.Join("cache", hashFileName)

	oldHash, _ := ioutil.ReadFile(cacheFilePath)
	oldHashStr := string(oldHash)

	if oldHashStr != currentHashStr {
		_ = ioutil.WriteFile(cacheFilePath, []byte(currentHashStr), 0644)
		return true
	}

	return false
}
