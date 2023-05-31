package structs

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/disgoorg/snowflake/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AnilistGraphQLQuery struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type AnilistMedia struct {
	ID           primitive.ObjectID     `bson:"_id,omitempty"`
	IdAnilist    int                    `json:"id" bson:"id_anilist"`
	IdMal        int                    `json:"idMal" bson:"id_mal"`
	Title        AnilistMediaTitle      `bson:"title"`
	Type         string                 `json:"type" bson:"type"`
	Format       string                 `json:"format" bson:"format"`
	Status       string                 `json:"status" bson:"status"`
	Description  string                 `json:"description" bson:"description"`
	StartDate    AnilistFuzzyDate       `bson:"start_date"`
	EndDate      AnilistFuzzyDate       `bson:"end_date"`
	Season       string                 `json:"season" bson:"season"`
	SeasonYear   int                    `json:"seasonYear" bson:"season_year"`
	SeasonInt    int                    `json:"seasonInt" bson:"season_int"`
	Episodes     int                    `bson:"episodes,omitempty"`
	Duration     int                    `bson:"duration,omitempty"`
	Chapters     int                    `bson:"chapters,omitempty"`
	Volumes      int                    `bson:"volumes,omitempty"`
	Source       string                 `json:"source" bson:"source"`
	Trailer      *AnilistMediaTrailer   `bson:"trailer,omitempty"`
	UpdatedAt    int                    `json:"updatedAt" bson:"updated_at"`
	CoverImage   AnilistMediaCoverImage `bson:"cover_image"`
	BannerImage  string                 `json:"bannerImage" bson:"banner_image"`
	Genres       []string               `bson:"genres"`
	Synonyms     []string               `bson:"synonyms"`
	AverageScore int                    `json:"averageScore" bson:"average_score"`
	MeanScore    int                    `json:"meanScore" bson:"mean_score"`
	Tags         []AnilistMediaTag      `bson:"tags"`
	IsAdult      bool                   `json:"isAdult" bson:"is_adult"`
	SiteUrl      string                 `json:"siteUrl" bson:"site_url"`
	MediaHash    string                 `bson:"media_hash"`
}

type AnilistMediaTitle struct {
	Romaji  string `bson:"romaji"`
	English string `bson:"english"`
	Native  string `bson:"native"`
}

type AnilistFuzzyDate struct {
	Year  int `bson:"year"`
	Month int `bson:"month"`
	Day   int `bson:"day"`
}

type AnilistMediaTrailer struct {
	ID        string `bson:"id"`
	Site      string `bson:"site"`
	Thumbnail string `bson:"thumbnail"`
}

type AnilistMediaCoverImage struct {
	ExtraLarge string `bson:"extra_large"`
	Large      string `bson:"large"`
	Medium     string `bson:"medium"`
	Color      string `bson:"color"`
}

type AnilistMediaTag struct {
	ID          int    `bson:"id"`
	Name        string `bson:"name"`
	Description string `bson:"description"`
	Category    string `bson:"category"`
	Rank        int    `bson:"rank"`
	IsAdult     bool   `json:"isAdult" bson:"is_adult"`
}

// Helper functions for Anilist Media
func (media *AnilistMedia) GetTitle() string {
	mediaTitle := strings.TrimSpace(media.Title.English)

	if mediaTitle == "" {
		mediaTitle = strings.TrimSpace(media.Title.Romaji)
	}

	if mediaTitle == "" {
		mediaTitle = strings.TrimSpace(media.Title.Native)
	}

	return mediaTitle
}

func (date *AnilistFuzzyDate) Hash() string {
	if date == nil {
		return ""
	}

	return strconv.Itoa(date.Year) + strconv.Itoa(date.Month) + strconv.Itoa(date.Day)
}

func (trailer *AnilistMediaTrailer) Hash() string {
	if trailer == nil {
		return ""
	}
	return trailer.ID + trailer.Site + trailer.Thumbnail
}

func (coverImage *AnilistMediaCoverImage) Hash() string {
	if coverImage == nil {
		return ""
	}

	return coverImage.ExtraLarge + coverImage.Large + coverImage.Medium + coverImage.Color
}

func (media *AnilistMedia) GetTags() string {
	// Create a copy of the tags slice to avoid modifying the original tags.
	tags := make([]AnilistMediaTag, len(media.Tags))
	copy(tags, media.Tags)

	// Sort the tags by rank.
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Rank < tags[j].Rank
	})

	// Select the top 5 tags.
	if len(tags) > 8 {
		tags = tags[:8]
	}

	// Extract the tag names and join them as a string.
	tagNames := make([]string, len(tags))
	for i, tag := range tags {
		tagNames[i] = tag.Name
	}

	return strings.Join(tagNames, ", ")
}

func (media *AnilistMedia) GetGenres() string {
	return strings.Join(media.Genres, ", ")
}

func (media *AnilistMedia) GetSynonyms() string {
	return strings.Join(media.Synonyms, ", ")
}

func (media *AnilistMedia) GetStartDate() string {
	return strconv.Itoa(media.StartDate.Year) + "-" + strconv.Itoa(media.StartDate.Month) + "-" + strconv.Itoa(media.StartDate.Day)
}

func (media *AnilistMedia) GetEndDate() string {
	return strconv.Itoa(media.EndDate.Year) + "-" + strconv.Itoa(media.EndDate.Month) + "-" + strconv.Itoa(media.EndDate.Day)
}

func (media *AnilistMedia) GetMeanScoreStr() string {
	return strconv.Itoa(media.MeanScore) + "%"
}

func (media *AnilistMedia) GetSeasonYearStr() string {
	return strconv.Itoa(media.SeasonYear)
}

func (media *AnilistMedia) GetEpisodesStr() string {
	return strconv.Itoa(media.Episodes)
}

func (media *AnilistMedia) GetChapterStr() string {
	return strconv.Itoa(media.Chapters)
}

func (media *AnilistMedia) GetVolumeStr() string {
	return strconv.Itoa(media.Volumes)
}

func (media *AnilistMedia) GetStatus() string {
	switch media.Status {
	case "FINISHED":
		return "Finished"
	case "RELEASING":
		return "Releasing"
	case "NOT_YET_RELEASED":
		return "Not yet released"
	case "CANCELLED":
		return "Cancelled"
	case "HIATUS":
		return "Hiatus"
	default:
		return "Unknown"
	}
}

func (media *AnilistMedia) GetSource() string {
	switch media.Source {
	case "ORIGINAL":
		return "Original"
	case "MANGA":
		return "Manga"
	case "LIGHT_NOVEL":
		return "Light novel"
	case "VISUAL_NOVEL":
		return "Visual novel"
	case "VIDEO_GAME":
		return "Video game"
	case "OTHER":
		return "Other"
	case "NOVEL":
		return "Novel"
	case "DOUJINSHI":
		return "Doujinshi"
	case "media":
		return "media"
	case "WEB_NOVEL":
		return "Web novel"
	case "LIVE_ACTION":
		return "Live action"
	case "GAME":
		return "Game"
	case "COMIC":
		return "Comic"
	case "MULTIMEDIA_PROJECT":
		return "Multimedia project"
	case "PICTURE_BOOK":
		return "Picture book"
	default:
		return "Unknown"
	}
}

func (media *AnilistMedia) GetFormat() string {
	switch media.Format {
	case "TV":
		return "TV"
	case "TV_SHORT":
		return "TV short"
	case "MOVIE":
		return "Movie"
	case "SPECIAL":
		return "Special"
	case "OVA":
		return "Original Video Animation (OVA)"
	case "ONA":
		return "Original Net Animation (ONA)"
	case "MUSIC":
		return "Music"
	case "MANGA":
		return "Manga"
	case "NOVEL":
		return "Novel"
	case "ONE_SHOT":
		return "One shot"
	default:
		return "Unknown"
	}
}

func (media *AnilistMedia) GetDescription() string {
	if len(media.Description) > 400 {
		return media.Description[:400] + "..."
	}
	return media.Description
}

func (media *AnilistMedia) Hash() string {
	var sb strings.Builder

	sb.WriteString(strconv.Itoa(media.IdMal))
	sb.WriteString(media.Title.Romaji)
	sb.WriteString(media.Title.English)
	sb.WriteString(media.Title.Native)
	sb.WriteString(string(media.Type))
	sb.WriteString(media.Format)
	sb.WriteString(media.Status)
	sb.WriteString(media.Description)
	sb.WriteString(media.StartDate.Hash())
	sb.WriteString(media.EndDate.Hash())
	sb.WriteString(media.Season)
	sb.WriteString(strconv.Itoa(media.SeasonYear))
	sb.WriteString(strconv.Itoa(media.SeasonInt))
	sb.WriteString(strconv.Itoa(media.Episodes))
	sb.WriteString(strconv.Itoa(media.Duration))
	sb.WriteString(media.Source)
	sb.WriteString(media.Trailer.Hash())
	sb.WriteString(strconv.Itoa(media.UpdatedAt))
	sb.WriteString(media.CoverImage.Hash())
	sb.WriteString(media.BannerImage)
	sb.WriteString(strings.Join(media.Genres, ","))
	sb.WriteString(strings.Join(media.Synonyms, ","))
	sb.WriteString(strconv.Itoa(media.AverageScore))
	sb.WriteString(strconv.Itoa(media.MeanScore))
	sb.WriteString(media.GetTags())
	sb.WriteString(strconv.FormatBool(media.IsAdult))
	sb.WriteString(media.SiteUrl)

	hasher := sha256.New()
	hasher.Write([]byte(sb.String()))
	return hex.EncodeToString(hasher.Sum(nil))
}

// Structs for storing media search query for maintaining state
// for discord embed navigation
type AnilistSearchQuery struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	SearchText string             `bson:"search_text"`
	MediaType  string             `bson:"media_type"`
	CreatedAt  time.Time          `bson:"created_at,omitempty"`
	LastUsedAt time.Time          `bson:"last_used_at,omitempty"`
}

// User media tracking
type UserMedia struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	MediaID   primitive.ObjectID `bson:"media_id"`
	MediaType string             `bson:"media_type"`
	UserID    snowflake.ID       `bson:"user_id"`
	Status    string             `bson:"status"`
	Score     int                `bson:"score"`
	CreatedAt time.Time          `bson:"created_at,omitempty"`
	UpdatedAt time.Time          `bson:"updated_at,omitempty"`
}

func (userMedia *UserMedia) GetStatus() string {
	currentStatus := "Watching"
	repeatingStatus := "Repeating"

	if userMedia.MediaType == "MANGA" {
		currentStatus = "Reading"
		repeatingStatus = "Rereading"
	}

	switch userMedia.Status {
	case "CURRENT":
		return currentStatus
	case "PLANNING":
		return "Planning"
	case "COMPLETED":
		return "Completed"
	case "DROPPED":
		return "Dropped"
	case "PAUSED":
		return "Paused"
	case "REPEATING":
		return repeatingStatus
	default:
		return "Unknown"
	}
}

func (userMedia *UserMedia) GetType() string {
	switch userMedia.MediaType {
	case "ANIME":
		return "Anime"
	case "MANGA":
		return "Manga"
	default:
		return "Unknown"
	}
}
