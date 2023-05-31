package anilist

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

type AnimeOfflineEntry struct {
	Title    string   `json:"title"`
	Synonyms []string `json:"synonyms"`
	Hash     string   `json:"hash"`
}

type AnimeOfflineData struct {
	License    map[string]string   `json:"license"`
	Repository string              `json:"repository"`
	LastUpdate string              `json:"lastUpdate"`
	Data       []AnimeOfflineEntry `json:"data"`
}

type animeOfflineBranchInfo struct {
	Commit struct {
		SHA string `json:"sha"`
	} `json:"commit"`
}

const ANIME_OFFLINE_SEARCH_RESULTS = 10

var animeOfflineEntriesMutex sync.RWMutex

// Stores anime details
var animeOfflineEntries []AnimeOfflineEntry

func InitializeAnime() {
	log.Println("AnimeOffline: Initializing database")
	loadAnimeOfflineDb()
	go syncAnimeOfflineDb()

	go func() {
		ticker := time.NewTicker(12 * time.Hour)
		for range ticker.C {
			syncAnimeOfflineDb()
		}
	}()
}

func SearchAnimeOfflineDb(searchText string) []AnimeOfflineEntry {
	animeOfflineEntriesMutex.RLock()
	defer animeOfflineEntriesMutex.RUnlock()

	if animeOfflineEntries == nil {
		return []AnimeOfflineEntry{}
	}

	var results []AnimeOfflineEntry
	searchText = strings.ToLower(searchText)

	// Check if there are Anime starting with the search text
	for _, entry := range animeOfflineEntries {
		if strings.HasPrefix(strings.ToLower(entry.Title), searchText) {
			results = append(results, entry)
			if len(results) >= ANIME_OFFLINE_SEARCH_RESULTS {
				return results
			}
			continue
		}

		for _, synonym := range entry.Synonyms {
			if strings.HasPrefix(strings.ToLower(synonym), searchText) {
				results = append(results, entry)
				if len(results) >= ANIME_OFFLINE_SEARCH_RESULTS {
					return results
				}
				break
			}
		}

		// Check if there are Anime containing the search text
		if strings.Contains(strings.ToLower(entry.Title), searchText) {
			results = append(results, entry)
			if len(results) >= ANIME_OFFLINE_SEARCH_RESULTS {
				return results
			}
			continue
		}

		for _, synonym := range entry.Synonyms {
			if strings.Contains(strings.ToLower(synonym), searchText) {
				results = append(results, entry)
				if len(results) >= ANIME_OFFLINE_SEARCH_RESULTS {
					return results
				}
				break
			}
		}

		// Break if we have enough results
		if len(results) >= ANIME_OFFLINE_SEARCH_RESULTS {
			return results
		}
	}

	return results
}

// Download the new database and update the in-memory database
func syncAnimeOfflineDb() {
	if !checkAnimeOfflineDbOutOfDate() {
		return
	}
	downloadAnimeOfflineDb()
	loadAnimeOfflineDb()
}

/*
* Loads the anime offline database to memory
 */
func loadAnimeOfflineDb() {
	// Acquire write lock to update the slice
	animeOfflineEntriesMutex.Lock()
	defer animeOfflineEntriesMutex.Unlock()

	jsonFile, err := os.Open(path.Join("cache", "anime-offline-database.json"))

	if err != nil {
		log.Println("AnimeOffline: Failed to open anime offline database:", err)
		return
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var animeOfflineData AnimeOfflineData

	err = json.Unmarshal(byteValue, &animeOfflineData)

	if err != nil {
		log.Println("AnimeOffline: Failed to parse anime offline database:", err)
		return
	}

	animeOfflineEntries = make([]AnimeOfflineEntry, 0)

	for _, entry := range animeOfflineData.Data {
		entry.Hash = computeHash(entry)
		animeOfflineEntries = append(animeOfflineEntries, entry)
	}

	// log first entry (0th entry)
	log.Printf("AnimeOffline: Loaded %d anime entries", len(animeOfflineEntries))
}

/*
* Check if the anime offline database is out of date
* Returns true if the database needs to be updated
 */
func checkAnimeOfflineDbOutOfDate() bool {
	log.Println("AnimeOffline: Checking if database is out of date")

	url := "https://api.github.com/repos/manami-project/anime-offline-database/branches/master"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error making request:", err)
		return false
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return false
	}

	var branchInfo animeOfflineBranchInfo
	err = json.Unmarshal(body, &branchInfo)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return false
	}

	sha := branchInfo.Commit.SHA

	cacheFilePath := path.Join("cache", "anime-offline-database-version")
	cachedSha := ""

	if _, err := os.Stat(cacheFilePath); err == nil {
		cachedShaBytes, err := ioutil.ReadFile(cacheFilePath)
		if err != nil {
			fmt.Println("Error reading file:", err)
			return false
		}
		cachedSha = string(cachedShaBytes)
	}

	err = ioutil.WriteFile(cacheFilePath, []byte(sha), 0644)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return false
	}

	// log if the database is out of date
	if sha != cachedSha {
		log.Println("AnimeOffline: Database is out of date")
	} else {
		log.Println("AnimeOffline: Database is up to date")
	}

	return sha != cachedSha
}

func downloadAnimeOfflineDb() {
	url := "https://github.com/manami-project/anime-offline-database/raw/master/anime-offline-database-minified.json"
	filePath := path.Join("cache", "anime-offline-database.json")

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: HTTP status code %d\n", resp.StatusCode)
		return
	}

	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	fmt.Printf("File %s downloaded successfully.\n", filePath)
}

func computeHash(entry AnimeOfflineEntry) string {
	hashData := entry.Title + strings.Join(entry.Synonyms, "")
	hash := sha256.Sum256([]byte(hashData))
	return hex.EncodeToString(hash[:])
}
