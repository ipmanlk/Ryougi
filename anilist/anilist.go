package anilist

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"ipmanlk/saika/database"
	"ipmanlk/saika/structs"
	"log"
	"net/http"
	"regexp"
)

func SearchMedia(searchText string, mediaType string) ([]structs.AnilistMedia, error) {
	apiURL := "https://graphql.anilist.co"

	query := `
	query ($search: String, $type: MediaType) {
		Page(perPage: 20) {
			media(search: $search, type: $type) {
				id
				idMal
				title {
					romaji
					english
					native
				}
				type
				format
				status
				description(asHtml: false)
				startDate {
					year
					month
					day
				}
				endDate {
					year
					month
					day
				}
				season
				seasonYear
				seasonInt
				episodes
				chapters
				volumes
				duration
				source
				trailer {
					id
					site
					thumbnail
				}
				updatedAt
				coverImage {
					extraLarge
					large
					medium
					color
				}
				bannerImage
				genres
				synonyms
				averageScore
				meanScore
				tags {
					id
					name
					description
					category
					rank
					isGeneralSpoiler
					isMediaSpoiler
					isAdult
					userId
				}
				isAdult
				siteUrl
			}
		}
	}`

	variables := map[string]interface{}{"search": searchText, "type": mediaType}

	reqBody := structs.AnilistGraphQLQuery{Query: query, Variables: variables}
	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(reqJSON))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data struct {
			Page struct {
				Media []structs.AnilistMedia
			}
		}
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	// Remove HTML tags from descriptions
	for i, media := range result.Data.Page.Media {
		result.Data.Page.Media[i].Description = removeHTMLTags(media.Description)
	}

	// save all media in a new goroutine, handle errors in the background
	err = database.SaveMedia(result.Data.Page.Media)
	if err != nil {
		log.Printf("Error saving media: %v", err)
	}

	return result.Data.Page.Media, nil
}

func removeHTMLTags(input string) string {
	// Use a regular expression to match HTML tags
	re := regexp.MustCompile("<[^>]*>")
	// Replace HTML tags with an empty string
	return re.ReplaceAllString(input, "")
}
