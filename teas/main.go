package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"gitlab.com/hokiegeek.net/teadb"
)

func main() {
	// Command flags: load
	loadCommand := flag.NewFlagSet("load", flag.ExitOnError)
	loadFilePtr := loadCommand.String("file", "", "The filename to use")
	isSpreadsheetPtr := loadCommand.Bool("spreadsheet", false, "Set if it should be processed as a spreadsheet")

	// Command flags: save
	saveCommand := flag.NewFlagSet("save", flag.ExitOnError)
	saveFilePtr := saveCommand.String("file", "", "The filename to use")

	command := os.Args[1]

	switch command {
	case "load":
		loadCommand.Parse(os.Args[2:])
		if *isSpreadsheetPtr {
			if err := loadFromSpreadsheetJSON(*loadFilePtr); err != nil {
				panic(err)
			}
		} else {
			if err := loadFromJSON(*loadFilePtr); err != nil {
				panic(err)
			}
		}
	case "save":
		saveCommand.Parse(os.Args[2:])
		if err := saveToFile(*saveFilePtr); err != nil {
			panic(err)
		}
	default:
		log.Fatalf("Unrecognized command: %s\n", command)
	}
}

func loadFromSpreadsheetJSON(filename string) error {
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	var steas []spreadsheetTea
	json.Unmarshal(raw, &steas)

	// fmt.Printf("%v\n", steas)

	var teas []teadb.Tea
	for _, stea := range steas {
		fmt.Printf("sTea (%s): %s\n", stea.ID, stea.Name)
		teas = append(teas, convertSpreadsheetTeaToTea(stea))
	}

	return createTeas(teas)
}

func loadFromJSON(filename string) error {
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	var teas []teadb.Tea
	json.Unmarshal(raw, &teas)

	return createTeas(teas)
}

func createTeas(teas []teadb.Tea) error {
	// fmt.Printf("%v\n", teas)

	for _, tea := range teas {
		// fmt.Printf("Tea (%d): %q, %s, %d, %s, %s, %s, %s\n", tea.ID, tea.Name, tea.Country, len(tea.Entries), tea.Packaging, tea.Flush, tea.Purchasedate.Format("1/2/2006"), tea.Date.Format("1/2/2006"))
		// fmt.Printf("Tea (%d): %q, %s\n", tea.ID, tea.Name, tea.Purchasedate.Format("1/2/2006"))
		// fmt.Printf("Tea (%d): %s\n", tea.ID, tea.Name)
		/*
			for _, entry := range tea.Entries {
				// fmt.Printf("  %s@%d: %s\n", entry.Date, entry.Time, entry.Datetime.Format("1/02/2006@1504"))
				fmt.Printf("  %s, %v\n", entry.Datetime.Format("1/02/2006@1504"), entry.Fixins)
			}
		*/
		if err := teadb.CreateTea(tea); err != nil {
			fmt.Printf("Could not create tea %d: %v\n", tea.ID, err)
		} else {
			fmt.Printf("Tea (%d): %s\n", tea.ID, tea.Name)
		}
		/*
		 */
	}

	return nil
}

type spreadsheetTea struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Timestamp        string   `json:"timestamp"` // TODO
	Date             string   `json:"date"`      // TODO
	Type             string   `json:"type"`
	Region           string   `json:"region"`
	Year             string   `json:"year"`
	FlushIdx         string   `json:"flush_idx"`
	Purchaselocation string   `json:"purchaselocation"`
	Purchasedate     string   `json:"purchasedate"`
	Purchaseprice    string   `json:"purchaseprice"`
	Comments         string   `json:"comments"`
	Pictures         []string `json:"pictures"`
	Country          string   `json:"country"`
	Leafgrade        string   `json:"leafgrade"`
	Blendedteas      string   `json:"blendedteas"`
	Blendratio       string   `json:"blendratio"`
	Size             string   `json:"size"`
	Stocked          bool     `json:"stocked"`
	Aging            bool     `json:"aging"`
	PackagingIdx     string   `json:"packaging_idx"`
	Sample           bool     `json:"sample"`
	Entries          []struct {
		Comments          string   `json:"comments"`
		Timestamp         string   `json:"timestamp"`
		Datetime          string   `json:"datetime"`
		Rating            string   `json:"rating"`
		Pictures          []string `json:"pictures"`
		Steeptime         string   `json:"steeptime"`
		SteepingvesselIdx string   `json:"steepingvessel_idx"`
		Steeptemperature  string   `json:"steeptemperature"` // TODO: in F
		Sessioninstance   string   `json:"sessioninstance"`
		Sessionclosed     bool     `json:"sessionclosed"`
		FixinsList        []int    `json:"fixins_list"`
	} `json:"entries"`
}

func convertSpreadsheetTeaToTea(sTea spreadsheetTea) teadb.Tea {
	teaPackagingTypes := []string{"Loose Leaf", "Bagged", "Tuo", "Beeng", "Brick", "Mushroom", "Square"}
	teaFlushTypesDefault := []string{"Spring", "Summer", "Fall", "Winter"}
	teaFlushTypesIndian := []string{"1st Flush", "2nd Flush", "Monsoon Flush", "Autumn Flush"}

	teaFixins := []string{"Milk", "Cream", "Half & Half", "Sugar", "Brown Sugar", "Raw Sugar",
		"Honey", "Vanilla Extract", "Vanilla Bean", "Maple Cream", "Maple Sugar", "Chai Goop", "Ice"}

	var tea teadb.Tea
	sTeaJSON, _ := json.Marshal(sTea)

	var err error

	tea.ID, _ = strconv.Atoi(sTea.ID)
	tea.Name = sTea.Name
	tea.Timestamp = sTea.Timestamp
	if dummyTime, err := time.Parse("1/2/2006", sTea.Date); err != nil {
		fmt.Printf("ERROR: Could not parse date: %s\n %s\n", err, sTeaJSON)
	} else {
		tea.Date = &dummyTime
	}
	tea.Type = sTea.Type
	tea.Region = sTea.Region
	tea.Year, _ = strconv.Atoi(sTea.Year)
	if len(sTea.FlushIdx) > 0 {
		flushIdx, err := strconv.Atoi(sTea.FlushIdx)
		if err != nil {
			fmt.Printf("ERROR: Could not process flush_idx: %s\n%s\n", err, sTeaJSON)
		}
		if "india" == strings.ToLower(sTea.Country) {
			tea.Flush = teaFlushTypesIndian[flushIdx]
		} else {
			tea.Flush = teaFlushTypesDefault[flushIdx]
		}
	}

	tea.Purchaselocation = sTea.Purchaselocation
	if len(sTea.Purchasedate) > 0 {
		if dummyTime, err := time.Parse("1/2/2006", sTea.Purchasedate); err != nil {
			fmt.Printf("ERROR: Could not parse purchasedate: %s\n %s\n", err, sTeaJSON)
		} else {
			tea.Purchasedate = &dummyTime
		}
	}
	tea.Purchaseprice, _ = strconv.Atoi(sTea.Purchaseprice)
	tea.Comments = sTea.Comments
	tea.Pictures = sTea.Pictures
	tea.Country = sTea.Country
	tea.Leafgrade = sTea.Leafgrade
	tea.Blendedteas = sTea.Blendedteas
	tea.Blendratio = sTea.Blendratio
	tea.Size = sTea.Size
	tea.Stocked = sTea.Stocked
	tea.Aging = sTea.Aging
	packagingIdx, err := strconv.Atoi(sTea.PackagingIdx)
	if err != nil {
		fmt.Printf("ERROR: Could not process packaging_idx: %d\n", packagingIdx)
	}
	tea.Packaging = teaPackagingTypes[packagingIdx]
	tea.Sample = sTea.Sample

	for _, sentry := range sTea.Entries {
		sentryJSON, _ := json.Marshal(sentry)
		var entry teadb.TeaEntry
		entry.Comments = sentry.Comments
		entry.Timestamp = sentry.Timestamp

		// if dummyTime, err := time.Parse("2006-01-02T15:04:00.000Z", sentry.Datetime); err != nil {
		if dummyTime, err := time.Parse(time.RFC3339Nano, sentry.Datetime); err != nil {
			fmt.Printf("ERROR: Could not parse Datetime: %s\n %s\n", err, sTeaJSON)
		} else {
			entry.Datetime = &dummyTime
		}

		entry.Rating, err = strconv.Atoi(sentry.Rating)
		if err != nil {
			fmt.Printf("ERROR: Could not process entry Rating: %s\n", sentryJSON)
		}
		entry.Steeptime = sentry.Steeptime
		entry.SteepingvesselIdx, err = strconv.Atoi(sentry.SteepingvesselIdx)
		if err != nil {
			fmt.Printf("ERROR: Could not process entry SteepingvesselIdx: %s\n", sentryJSON)
		}
		entry.Steeptemperature, err = strconv.Atoi(sentry.Steeptemperature)
		if err != nil {
			fmt.Printf("ERROR: Could not process entry Steeptemperature: %s\n", sentryJSON)
		}
		entry.Sessioninstance = sentry.Sessioninstance
		entry.Sessionclosed = sentry.Sessionclosed

		for _, pic := range sentry.Pictures {
			entry.Pictures = append(entry.Pictures, pic)
		}

		for _, fixin := range sentry.FixinsList {
			entry.Fixins = append(entry.Fixins, teaFixins[fixin])
		}

		tea.Entries = append(tea.Entries, entry)
	}

	return tea
}

func saveToFile(filename string) error {
	teas, err := teadb.GetAllTeas()
	if err != nil {
		return err
	}

	fmt.Printf("Saving %d teas\n", len(teas))

	teasJSON, err := json.Marshal(teas)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, teasJSON, 0644)
}
