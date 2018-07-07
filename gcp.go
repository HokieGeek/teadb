package teadb

import (
	"fmt"
	"log"
	"strconv"
	"time"

	// Imports the Google Butt Datastore client package.
	"cloud.google.com/go/datastore"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
)

const kindTea = "tea"

// TeaEntry encapsulates the data needed for a journal entry
type TeaEntry struct {
	Comments          string     `json:"comments"`
	Timestamp         string     `json:"timestamp"`
	Datetime          *time.Time `json:"datetime"`
	Rating            int        `json:"rating"`
	Pictures          []string   `json:"pictures"`
	Steeptime         string     `json:"steeptime"`
	SteepingvesselIdx int        `json:"steepingvessel_idx"`
	Steeptemperature  int        `json:"steeptemperature"` // TODO: in F
	Sessioninstance   string     `json:"sessioninstance"`
	Sessionclosed     bool       `json:"sessionclosed"`
	Fixins            []string   `json:"fixins"`
}

// Tea encapsulates a specific tea and its journal entries
type Tea struct {
	ID               int        `json:"id"`
	Name             string     `json:"name"`
	Timestamp        string     `json:"timestamp"` // TODO
	Date             *time.Time `json:"date"`      // TODO
	Type             string     `json:"type"`
	Region           string     `json:"region"`
	Year             int        `json:"year"`
	Flush            string     `json:"flush"`
	Purchaselocation string     `json:"purchaselocation"`
	Purchasedate     *time.Time `json:"purchasedate"`
	Purchaseprice    float32    `json:"purchaseprice"`
	Comments         string     `json:"comments"`
	Pictures         []string   `json:"pictures"`
	Country          string     `json:"country"`
	Leafgrade        string     `json:"leafgrade"`
	Blendedteas      string     `json:"blendedteas"`
	Blendratio       string     `json:"blendratio"`
	Size             string     `json:"size"`
	Stocked          bool       `json:"stocked"`
	Aging            bool       `json:"aging"`
	Packaging        string     `json:"packaging"`
	Sample           bool       `json:"sample"`
	Entries          []TeaEntry `json:"entries"`
}

func getClient(ctx context.Context) (*datastore.Client, error) {
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
func CreateTea(tea Tea) error {
	// TODO: validate?
	return saveTea(tea)
}

// UpdateTea updates a new tea entity
func UpdateTea(tea Tea) error {
	// TODO: validate?
	return saveTea(tea)
}

// CreateEntry creates a new entry on an existing tea
func CreateEntry(id int, entry TeaEntry) error {
	log.Printf("createEntry(%d): %v\n", id, entry)
	tea, err := GetTeaByID(id)
	if err != nil {
		return err
	}

	// TODO: validate
	tea.Entries = append(tea.Entries, entry)

	return saveTea(tea)
}

// UpdateEntry updates an existing entry
func UpdateEntry(id int, entry TeaEntry) error {
	tea, err := GetTeaByID(id)
	if err != nil {
		return err
	}

	// TODO: validate

	for i, teaEntry := range tea.Entries {
		if entry.Datetime.UnixNano() == teaEntry.Datetime.UnixNano() {
			tea.Entries[i] = entry
			break
		}
	}

	return saveTea(tea)
}

// GetAllTeas retrieves every tea entity
func GetAllTeas() ([]Tea, error) {
	ctx := context.Background()
	client, err := getClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	query := datastore.NewQuery(kindTea)

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

	key := datastore.NameKey(kindTea, strconv.Itoa(id), nil)

	var t Tea
	if err = client.Get(ctx, key, &t); err != nil {
		return t, err
	}

	return t, nil
}

func saveTea(tea Tea) error {
	ctx := context.Background()
	client, err := getClient(ctx)
	if err != nil {
		return fmt.Errorf("Failed to create client: %v", err)
	}

	// Create the Key instance.
	key := datastore.NameKey(kindTea, strconv.Itoa(tea.ID), nil)

	// Saves the new entity.
	if _, err := client.Put(ctx, key, &tea); err != nil {
		return fmt.Errorf("Failed to save tea: %v", err)
	}

	// fmt.Printf("Saved %v\n", key)
	return nil
}
