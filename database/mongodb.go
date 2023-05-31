package database

import (
	"context"
	"errors"
	"ipmanlk/saika/config"
	"ipmanlk/saika/structs"
	"log"
	"sync"
	"time"

	"github.com/disgoorg/snowflake/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	clientInstance *mongo.Client
	once           sync.Once
	database       = config.GetEnv("MONGO_DATABASE", "")
)

func InitMongoDb() {
	ensureUniqueIndex(GetCollection("media"), "media_hash")
}

func GetMongoClient() *mongo.Client {
	once.Do(func() {
		client, err := mongo.NewClient(options.Client().ApplyURI(config.GetEnv("MONGO_URI", "")))
		if err != nil {
			log.Fatalf("Failed to create client: %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err = client.Connect(ctx)
		if err != nil {
			log.Fatalf("Failed to connect client: %v", err)
		}

		clientInstance = client
	})

	return clientInstance
}

func DisconnectMongoClient(client *mongo.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client.Disconnect(ctx)
}

func GetCollection(collectionName string) *mongo.Collection {
	client := GetMongoClient()
	db := client.Database(database)
	return db.Collection(collectionName)
}

func ensureUniqueIndex(collection *mongo.Collection, fieldName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexModel := mongo.IndexModel{
		Keys:    bson.M{fieldName: 1},
		Options: options.Index().SetUnique(true),
	}

	_, err := collection.Indexes().CreateOne(ctx, indexModel)

	if err != nil {
		return err
	}

	return nil
}

func SaveMedia(media []structs.AnilistMedia) error {
	collection := GetCollection("media")
	// loop through the media.
	// 1. if id_anilist is not in the database, insert a new document
	// 2. if id_anilist is in the database and media_hash is different, update the document
	for _, media := range media {
		media.MediaHash = media.Hash()

		// check if the document exists in the database
		var result structs.AnilistMedia
		err := collection.FindOne(context.Background(), bson.M{"id_anilist": media.IdAnilist}).Decode(&result)
		if err != nil {
			if err != mongo.ErrNoDocuments {
				return err
			}
			// if the document does not exist, insert a new document
			_, err := collection.InsertOne(context.Background(), media)
			if err != nil {
				return err
			}
		} else {
			// if the document exists, check if the media_hash is different
			if result.Hash() != media.MediaHash {
				// if the media_hash is different, update the document
				_, err := collection.UpdateOne(context.Background(), bson.M{"id_anilist": media.IdAnilist}, bson.M{"$set": media})
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func GetMediaByIDAnilist(idAnilist int) (*structs.AnilistMedia, error) {
	collection := GetCollection("media")

	var result structs.AnilistMedia
	err := collection.FindOne(context.Background(), bson.M{"id_anilist": idAnilist}).Decode(&result)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func GetMediaByObjectID(objectID primitive.ObjectID) (*structs.AnilistMedia, error) {
	collection := GetCollection("media")

	var result structs.AnilistMedia
	err := collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&result)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func GetMediaByHexID(hexID string) (*structs.AnilistMedia, error) {
	objectID, err := primitive.ObjectIDFromHex(hexID)
	if err != nil {
		return nil, err
	}

	return GetMediaByObjectID(objectID)
}

func GetUserMediaByHexID(hexID string) (*structs.UserMedia, error) {
	objectID, err := primitive.ObjectIDFromHex(hexID)
	if err != nil {
		return nil, err
	}

	collection := GetCollection("user_media")

	var result structs.UserMedia
	err = collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&result)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

// function for saving anilist search query in the database
func SaveSearchQuery(storeQuery *structs.AnilistSearchQuery) (query *structs.AnilistSearchQuery, err error) {
	collection := GetCollection("search_queries")

	// check if the document exists in the database using user_id and search_text
	var result structs.AnilistSearchQuery

	err = collection.FindOne(context.Background(), bson.M{"search_text": storeQuery.SearchText, "media_type": storeQuery.MediaType}).Decode(&result)

	if err != nil {
		if err != mongo.ErrNoDocuments {
			return nil, err
		}

		storeQuery.CreatedAt = time.Now()
		storeQuery.LastUsedAt = time.Now()

		// if the document does not exist, insert a new document and return the document from the database
		// with mongodb id
		res, err := collection.InsertOne(context.Background(), storeQuery)
		if err != nil {
			return nil, err
		}

		err = collection.FindOne(context.Background(), bson.M{"_id": res.InsertedID}).Decode(&result)
		if err != nil {
			return nil, err
		}

		return &result, nil
	}

	// if document exists, update last_used_at and return the document from the database
	// with mongodb id
	_, err = collection.UpdateOne(context.Background(), bson.M{"_id": result.ID}, bson.M{"$set": bson.M{"last_used_at": time.Now()}})

	if err != nil {
		return nil, err
	}

	// document already exists
	return &result, nil
}

func GetSearchQueryByHexID(hexID string) (*structs.AnilistSearchQuery, error) {
	objectID, _ := primitive.ObjectIDFromHex(hexID)

	collection := GetCollection("search_queries")

	var result structs.AnilistSearchQuery
	err := collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&result)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func SearchMedia(searchText string, mediaType string, nsfw bool) ([]structs.AnilistMedia, error) {
	collection := GetCollection("media")

	// Regular expression to allow for case-insensitive search
	searchRegex := primitive.Regex{Pattern: searchText, Options: "i"}

	titleFilter := bson.M{
		"$and": []bson.M{
			{
				"$or": []bson.M{
					{"title.english": searchRegex},
					{"title.romaji": searchRegex},
					{"title.native": searchRegex},
				},
			},
			{"type": mediaType},
		},
	}

	if !nsfw {
		titleFilter["$and"] = append(titleFilter["$and"].([]bson.M), bson.M{"is_adult": false})
	}

	options := options.Find()
	options.SetLimit(20)

	// Find matches in the title first
	titleCursor, err := collection.Find(context.Background(), titleFilter, options)
	if err != nil {
		return nil, err
	}

	// decode the cursor to a slice of structs
	var titleMedia []structs.AnilistMedia
	err = titleCursor.All(context.Background(), &titleMedia)
	if err != nil {
		return nil, err
	}

	// If there are less than 20 matches, fill in the remaining slots with matches from other fields
	if len(titleMedia) < 20 {
		options.SetLimit(20 - int64(len(titleMedia)))

		// Get the ObjectIDs of the title matches
		var titleIDs []primitive.ObjectID
		for _, media := range titleMedia {
			titleIDs = append(titleIDs, media.ID)
		}

		// Query for matches in other fields, excluding the ones already matched by title
		otherFilter := bson.M{
			"$and": []bson.M{
				{
					"$or": []bson.M{
						{"description": searchRegex},
						{"genres": searchRegex},
						{"tags.name": searchRegex},
					},
				},
				{"type": mediaType},
				{"_id": bson.M{"$nin": titleIDs}},
			},
		}

		if !nsfw {
			otherFilter["$and"] = append(otherFilter["$and"].([]bson.M), bson.M{"is_adult": false})
		}

		otherCursor, err := collection.Find(context.Background(), otherFilter, options)
		if err != nil {
			return nil, err
		}

		var otherMedia []structs.AnilistMedia
		err = otherCursor.All(context.Background(), &otherMedia)
		if err != nil {
			return nil, err
		}

		// Concatenate the two slices
		titleMedia = append(titleMedia, otherMedia...)
	}

	return titleMedia, nil
}

func SaveUserMedia(userMedia *structs.UserMedia) (*structs.UserMedia, error) {
	collection := GetCollection("user_media")

	// check if the document exists in the database using user_id and search_text
	var result structs.UserMedia

	err := collection.FindOne(context.Background(), bson.M{"user_id": userMedia.UserID, "media_id": userMedia.MediaID}).Decode(&result)

	if err != nil {
		if err != mongo.ErrNoDocuments {
			return nil, err
		}

		if userMedia.Status == "" {
			userMedia.Status = "planning"
		}

		userMedia.CreatedAt = time.Now()
		userMedia.UpdatedAt = time.Now()

		// if the document does not exist, insert a new document and return the document from the database
		// with mongodb id
		res, err := collection.InsertOne(context.Background(), userMedia)
		if err != nil {
			return nil, err
		}

		err = collection.FindOne(context.Background(), bson.M{"_id": res.InsertedID}).Decode(&result)
		if err != nil {
			return nil, err
		}

		return &result, nil
	}

	// if document exists, update last_used_at and return the document from the database
	// with mongodb id
	var newStatus = userMedia.Status
	if newStatus == "" {
		newStatus = result.Status
	}

	var newScore = userMedia.Score
	if newScore == 0 {
		newScore = -1
	}

	var updateResult *mongo.UpdateResult
	updateResult, err = collection.UpdateOne(context.Background(), bson.M{"_id": result.ID}, bson.M{"$set": bson.M{"updated_at": time.Now(), "status": newStatus, "score": newScore}})

	if err != nil {
		return nil, err
	}

	if updateResult.ModifiedCount == 0 {
		return nil, errors.New("no documents were modified")
	}

	// return the updated document
	err = collection.FindOne(context.Background(), bson.M{"_id": result.ID}).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func GetAllUserMedia(userID snowflake.ID, mediaType string, status string) ([]structs.UserMedia, error) {
	collection := GetCollection("user_media")

	var filter bson.M
	if status == "" {
		filter = bson.M{
			"$and": []bson.M{
				{"user_id": userID},
				{"media_type": mediaType},
			},
		}
	} else {
		filter = bson.M{
			"$and": []bson.M{
				{"user_id": userID},
				{"media_type": mediaType},
				{"status": status},
			},
		}
	}

	options := options.Find()
	options.SetLimit(20)

	cursor, err := collection.Find(context.Background(), filter, options)
	if err != nil {
		return nil, err
	}

	// decode the cursor to a slice of structs
	var media []structs.UserMedia
	err = cursor.All(context.Background(), &media)
	if err != nil {
		return nil, err
	}

	return media, nil
}

func DeleteUserMediaByHexID(hexID string) error {
	objectID, _ := primitive.ObjectIDFromHex(hexID)

	collection := GetCollection("user_media")

	_, err := collection.DeleteOne(context.Background(), bson.M{"_id": objectID})

	if err != nil {
		return err
	}

	return nil
}
