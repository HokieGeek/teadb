package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"gitlab.com/hokiegeek.net/teadb"
)

func main() {
	// Command flags: load
	loadCommand := flag.NewFlagSet("load", flag.ExitOnError)
	loadFilePtr := loadCommand.String("file", "", "The filename to use")

	// Command flags: save
	saveCommand := flag.NewFlagSet("save", flag.ExitOnError)
	saveFilePtr := saveCommand.String("file", "", "The filename to use")

	command := os.Args[1]

	switch command {
	case "load":
		loadCommand.Parse(os.Args[2:])
		if err := loadFromFile(*loadFilePtr); err != nil {
			panic(err)
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

func loadFromFile(filename string) error {
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	var teas []spreadsheetTea
	json.Unmarshal(raw, &teas)

	for _, stea := range teas {
		tea := convertSpreadsheetTeaToTea(stea)
		// fmt.Printf("sTea (%s): %q, %d\n", stea.ID, stea.Name, len(stea.Entries))
		if err = teadb.CreateTea(tea); err != nil {
			fmt.Printf("Could not create tea %d: %v\n", tea.ID, err)
		} else {
			fmt.Printf("Tea (%d): %q, %d\n", tea.ID, tea.Name, len(tea.Entries))
		}
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
		Date              string   `json:"date"` // TODO
		Time              string   `json:"time"` // TODO
		Rating            string   `json:"rating"`
		Pictures          []string `json:"pictures"`
		Steeptime         string   `json:"steeptime"`
		SteepingvesselIdx string   `json:"steepingvessel_idx"`
		Steeptemperature  string   `json:"steeptemperature"` // TODO: in F
		Sessioninstance   string   `json:"sessioninstance"`
		Sessionclosed     bool     `json:"sessionclosed"`
		FixinsList        []string `json:"fixins_list"`
	} `json:"entries"`
}

func convertSpreadsheetTeaToTea(sTea spreadsheetTea) teadb.Tea {
	var tea teadb.Tea

	tea.ID, _ = strconv.Atoi(sTea.ID)
	tea.Name = sTea.Name
	tea.Timestamp = sTea.Timestamp
	tea.Date = sTea.Date
	tea.Type = sTea.Type
	tea.Region = sTea.Region
	tea.Year, _ = strconv.Atoi(sTea.Year)
	tea.FlushIdx, _ = strconv.Atoi(sTea.FlushIdx)
	tea.Purchaselocation = sTea.Purchaselocation
	tea.Purchasedate = sTea.Purchasedate
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
	tea.PackagingIdx, _ = strconv.Atoi(sTea.PackagingIdx)
	tea.Sample = sTea.Sample

	for _, sentry := range sTea.Entries {
		var entry teadb.TeaEntry
		entry.Comments = sentry.Comments
		entry.Timestamp = sentry.Timestamp
		entry.Date = sentry.Date
		entry.Time, _ = strconv.Atoi(sentry.Time)
		entry.Rating, _ = strconv.Atoi(sentry.Rating)
		entry.Steeptime = sentry.Steeptime
		entry.SteepingvesselIdx, _ = strconv.Atoi(sentry.SteepingvesselIdx)
		entry.Steeptemperature, _ = strconv.Atoi(sentry.Steeptemperature)
		entry.Sessioninstance = sentry.Sessioninstance
		entry.Sessionclosed = sentry.Sessionclosed

		for _, pic := range sentry.Pictures {
			entry.Pictures = append(entry.Pictures, pic)
		}

		for _, fixinStr := range sentry.FixinsList {
			fixin, _ := strconv.Atoi(fixinStr)
			entry.FixinsList = append(entry.FixinsList, fixin)
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

	teasJSON, err := json.Marshal(teas)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, teasJSON, 0644)
}
