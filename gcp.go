package teadb

import (
	"fmt"
	"strconv"
	"time"

	// Imports the Google Butt Datastore client package.
	"cloud.google.com/go/datastore"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
)

const kindTea = "tea"
const projectID = "hgnet-tea"

// TeaEntry encapsulates the data needed for a journal entry
type TeaEntry struct {
	Comments          string     `json:"comments"`
	Timestamp         string     `json:"timestamp"`
	Datetime          *time.Time `json:"datetime"`
	Rating            int        `json:"rating"`
	Pictures          []string   `json:"pictures"`
	Steeptime         int        `json:"steeptime"`
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
	Size             int        `json:"size"`
	Stocked          bool       `json:"stocked"`
	Aging            bool       `json:"aging"`
	Packaging        string     `json:"packaging"`
	Sample           bool       `json:"sample"`
	Entries          []TeaEntry `json:"entries"`
}

// GcpClient is the client struct
type GcpClient struct {
	ctx    context.Context
	client *datastore.Client
}

// CreateTea creates a new tea entity
func (c *GcpClient) CreateTea(tea Tea) error {
	// TODO: validate?
	return c.saveTea(tea)
}

// UpdateTea updates a new tea entity
func (c *GcpClient) UpdateTea(tea Tea) error {
	// TODO: validate?
	return c.saveTea(tea)
}

// DeleteTea deletes an existing tea
func (c *GcpClient) DeleteTea(teaID int) error {
	// TODO: validate?
	return c.removeTea(teaID)
}

// CreateEntry creates a new entry on an existing tea
func (c *GcpClient) CreateEntry(id int, entry TeaEntry) error {
	tea, err := c.GetTeaByID(id)
	if err != nil {
		return err
	}

	// TODO: validate
	tea.Entries = append(tea.Entries, entry)

	return c.saveTea(tea)
}

// UpdateEntry updates an existing entry
func (c *GcpClient) UpdateEntry(id int, entry TeaEntry) error {
	tea, err := c.GetTeaByID(id)
	if err != nil {
		return err
	}

	// TODO: validate

	found := false
	for i, teaEntry := range tea.Entries {
		if entry.Datetime.UnixNano() == teaEntry.Datetime.UnixNano() {
			found = true
			tea.Entries[i] = entry
			break
		}
	}

	if !found {
		return fmt.Errorf("Failed to find entry to update: %v", err)
	}

	return c.saveTea(tea)
}

// GetAllTeas retrieves every tea entity
func (c *GcpClient) GetAllTeas() ([]Tea, error) {
	query := datastore.NewQuery(kindTea)
	it := c.client.Run(c.ctx, query)

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
func (c *GcpClient) GetTeaByID(id int) (Tea, error) {
	key := datastore.NameKey(kindTea, strconv.Itoa(id), nil)

	var t Tea
	if err := c.client.Get(c.ctx, key, &t); err != nil {
		return t, err
	}

	return t, nil
}

func (c *GcpClient) saveTea(tea Tea) error {
	// Create the Key instance.
	key := datastore.NameKey(kindTea, strconv.Itoa(tea.ID), nil)

	// Saves the new entity.
	if _, err := c.client.Put(c.ctx, key, &tea); err != nil {
		return fmt.Errorf("Failed to save tea: %v", err)
	}

	// fmt.Printf("Saved %v\n", key)
	return nil
}

func (c *GcpClient) removeTea(teaID int) error {
	// Create the Key instance.
	key := datastore.NameKey(kindTea, strconv.Itoa(teaID), nil)

	// Saves the new entity.
	if err := c.client.Delete(c.ctx, key); err != nil {
		return fmt.Errorf("Failed to remove tea: %v", err)
	}

	// fmt.Printf("Saved %v\n", key)
	return nil
}

// New creates a new GcpClient
func New() (*GcpClient, error) {
	c := new(GcpClient)

	var err error
	c.ctx = context.Background()
	c.client, err = datastore.NewClient(c.ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("Failed to create client: %v", err)
	}

	return c, nil
}
