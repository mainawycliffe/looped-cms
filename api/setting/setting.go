package setting

import (
	"context"
	"fmt"

	"github.com/looped-dev/cms/api/constants"
	"github.com/looped-dev/cms/api/graph/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func NewSettingRepository(dbClient *mongo.Client) *SettingRepository {
	return &SettingRepository{
		DBClient: dbClient,
	}
}

type SettingRepository struct {
	DBClient *mongo.Client
}

// Details fetch the settings of the current settings from the database, returns
// nil if none is found.
func (setting *SettingRepository) Details(ctx context.Context) (*model.SiteSettings, error) {
	settings := &model.SiteSettings{}
	// TODO: figure out out to fetch the first record
	err := setting.DBClient.Database("cms").
		Collection(constants.SettingsCollectionName).
		FindOne(ctx, bson.M{}).
		Decode(settings)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("No settings found")
		}
		return nil, err
	}
	return settings, nil
}

// Exists checks whether settings have been set in the database
func (setting *SettingRepository) Exists(ctx context.Context) (bool, error) {
	count, err := setting.DBClient.Database("cms").
		Collection(constants.SettingsCollectionName).
		// count all documents, it should return 1 record as the collection is
		// capped to a single collection.
		CountDocuments(ctx, bson.D{})
	if err != nil {
		return false, err
	}
	// should only have a single document, as it is a capped collection
	return count == 1, nil
}

// SaveSettings saves the settings to the database. If the settings already, it
// updates existing settings, otherwise it creates a new settings. Also, it
// ensures only a single record will exist in the database.
func (setting *SettingRepository) SaveSettings(ctx context.Context, input model.SiteSettingsInput) (*model.SiteSettings, error) {
	_, err := setting.DBClient.Database("cms").
		Collection(constants.SettingsCollectionName).
		InsertOne(ctx, input)
	if err != nil {
		return nil, err
	}
	siteSettings := &model.SiteSettings{}
	err = setting.DBClient.Database("cms").
		Collection(constants.SettingsCollectionName).
		// using find one to get the first record as this collection is capped and
		// can only contain one record
		FindOne(ctx, bson.M{}).Decode(siteSettings)
	if err != nil {
		return nil, err
	}
	return siteSettings, nil
}
