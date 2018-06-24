package teadb

import (
	"fmt"
	"log"
	"strconv"

	// Imports the Google Butt Datastore client package.
	"cloud.google.com/go/datastore"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
)

const kind = "tea"

// Tea encapsulates a specific tea and its journal entries
type Tea struct {
	ID               int      `json:"id"`
	Name             string   `json:"name"`
	Timestamp        string   `json:"timestamp"` // TODO
	Date             string   `json:"date"`      // TODO
	Type             string   `json:"type"`
	Region           string   `json:"region"`
	Year             int      `json:"year"`
	FlushIdx         string   `json:"flush_idx"`
	Purchaselocation string   `json:"purchaselocation"`
	Purchasedate     string   `json:"purchasedate"`
	Purchaseprice    int      `json:"purchaseprice"`
	Comments         string   `json:"comments"`
	Pictures         []string `json:"pictures"`
	Country          string   `json:"country"`
	Leafgrade        string   `json:"leafgrade"`
	Blendedteas      string   `json:"blendedteas"`
	Blendratio       string   `json:"blendratio"`
	Size             string   `json:"size"`
	Stocked          bool     `json:"stocked"`
	Aging            bool     `json:"aging"`
	PackagingIdx     int      `json:"packaging_idx"`
	Sample           bool     `json:"sample"`
	Entries          []struct {
		Comments          string   `json:"comments"`
		Timestamp         string   `json:"timestamp"`
		Date              string   `json:"date"` // TODO
		Time              string   `json:"time"` // TODO
		Rating            int      `json:"rating"`
		Pictures          []string `json:"pictures"`
		Steeptime         string   `json:"steeptime"`
		SteepingvesselIdx int      `json:"steepingvessel_idx"`
		Steeptemperature  int      `json:"steeptemperature"` // TODO: in F
		Sessioninstance   string   `json:"sessioninstance"`
		Sessionclosed     bool     `json:"sessionclosed"`
		FixinsList        []int    `json:"fixins_list"`
	} `json:"entries"`
}

func getClient(ctx context.Context) (*datastore.Client, error) {
	// ctx := context.Background()

	// Set your Google Butt Platform project ID.
	projectID := "hgnet-tea"

	// Creates a client.
	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
		// log.Fatalf("Failed to create client: %v", err)
	}

	return client, nil
}

// CreateTea creates a new tea entity
func CreateTea(id int) {
	tea := Tea{
		ID:       id,
		Comments: "ADD MORE LATER",
	}

	saveTea(tea)
}

// CreateEntry creates a new entry on an existing tea
func CreateEntry() {
	// TODO
}

// GetAllTeas retrieves every tea entity
func GetAllTeas() ([]Tea, error) {
	ctx := context.Background()
	client, err := getClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	query := datastore.NewQuery(kind)

	it := client.Run(ctx, query)

	var teas []Tea

	for {
		var tea Tea
		_, err := it.Next(&tea)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("Error fetching next tea: %v", err)
		}

		teas = append(teas, tea)
		// fmt.Printf("Tea (%d): %q\n", tea.ID, tea.Comments)
	}

	return teas, nil
}

// GetTeaByID returns a single tea instance based on its ID
func GetTeaByID(id int) (Tea, error) {
	ctx := context.Background()
	client, err := getClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	key := datastore.NameKey(kind, strconv.Itoa(id), nil)

	var t Tea
	if err = client.Get(ctx, key, &t); err != nil {
		return t, err
	}

	return t, nil
}

func saveTea(tea Tea) {
	ctx := context.Background()
	client, err := getClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Create the Key instance.
	key := datastore.NameKey(kind, strconv.Itoa(tea.ID), nil)

	// Saves the new entity.
	if _, err := client.Put(ctx, key, &tea); err != nil {
		log.Fatalf("Failed to save tea: %v", err)
	}

	fmt.Printf("Saved %v\n", key)
}