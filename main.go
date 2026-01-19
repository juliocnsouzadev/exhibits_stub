package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type LocalizedString struct {
	En string `json:"en"`
	Ar string `json:"ar"`
}

type Coordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type ExhibitDTO struct {
    ID                   int             `json:"exhibit_id"`
	SiteName             LocalizedString `json:"site_name"`
	SiteBriefDescription LocalizedString `json:"site_brief_description"`
	NameEn               string          `json:"name_en"`
	NameAr               string          `json:"name_ar"`
	Name                 LocalizedString `json:"name"`
	BriefDescription     LocalizedString `json:"brief_description"`
	GeneratedDescription LocalizedString `json:"generated_description"`
	ArtistDescription    LocalizedString `json:"artist_description"`
	ArtistName           LocalizedString `json:"artist_name"`
	LocationDescription  LocalizedString `json:"location_description"`
	Type                 string          `json:"type"`
	ImageURL             string          `json:"image_url"`
	VideoURL             string          `json:"video_url"`
	RelevantLink         string          `json:"relevant_link"`
	AudioGuideLink       *string         `json:"audio_guide_link"`
	LocationURL          string          `json:"location_url"`
	Tags                 []string        `json:"tags"`
	Coords               Coordinates     `json:"coords"`
	Ownership            string          `json:"ownership"`
	RecreationLevel      string          `json:"recreation_level"`
}

// Artefact DTOs based on qm_schema.json
type Museum struct {
	Slug    string `json:"slug"`
	Label   string `json:"label"`
	LabelEN string `json:"labelEN"`
	LabelAR string `json:"labelAR"`
}

type Weekday struct {
	Number int    `json:"number"`
	Name   string `json:"name"`
}

type OpeningTime struct {
	OpeningAt string  `json:"openingAt"`
	ClosingAt string  `json:"closingAt"`
	Weekday   Weekday `json:"weekday"`
}

type FocalPoint struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type ObjectImage struct {
	URL         string     `json:"url"`
	Width       int        `json:"width"`
	Height      int        `json:"height"`
	FocalPoint  FocalPoint `json:"focalPoint"`
	AltTextEN   string     `json:"altTextEN"`
	AltTextAR   string     `json:"altTextAR"`
	CreditLineEN string    `json:"creditLineEN"`
	CreditLineAR string    `json:"creditLineAR"`
}

type ObjectImages struct {
	Original []ObjectImage `json:"original"`
	Card     []ObjectImage `json:"card"`
}

type ArtefactDTO struct {
	ObjectNumber   string        `json:"objectNumber"`
	TitleEN        string        `json:"titleEN"`
	TitleAR        string        `json:"titleAR"`
	ObjectNameEN   string        `json:"objectNameEN"`
	ObjectNameAR   string        `json:"objectNameAR"`
	ArtistEN       string        `json:"artistEN"`
	ArtistAR       string        `json:"artistAR"`
	Museum         Museum        `json:"museum"`
	OpeningTimes   []OpeningTime  `json:"openingTimes"`
	SummaryEN      string        `json:"summaryEN"`
	SummaryAR      string        `json:"summaryAR"`
	ObjectImages   ObjectImages  `json:"objectImages"`
	RelatedWebpages []interface{} `json:"relatedWebpages"`
	Object3dEmbed  []interface{} `json:"object3dEmbed"`
	Coords         Coordinates   `json:"coords"`
}

type QMResponse struct {
	Count    int           `json:"count"`
	Next     string        `json:"next"`
	Previous interface{}   `json:"previous"`
	Results  []ArtefactDTO `json:"results"`
}

// findFile tries to locate a file by checking multiple possible paths
func findFile(filename string) (string, error) {
	// Try current directory first
	if _, err := os.Stat(filename); err == nil {
		return filename, nil
	}

	// Try relative to executable
	execPath, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(execPath)
		path := filepath.Join(execDir, filename)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	// Try working directory
	wd, err := os.Getwd()
	if err == nil {
		path := filepath.Join(wd, filename)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	// Return original filename to get proper error message
	return filename, os.ErrNotExist
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Hello, world!"))
	})

	http.HandleFunc("/exhibits", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		filePath, err := findFile("exhibits.json")
		if err != nil {
			http.Error(w, "Error finding exhibits.json", http.StatusInternalServerError)
			log.Printf("Error finding file: %v", err)
			return
		}
		data, err := os.ReadFile(filePath)
		if err != nil {
			http.Error(w, "Error reading exhibits.json", http.StatusInternalServerError)
			log.Printf("Error reading file %s: %v", filePath, err)
			return
		}

		var exhibits []ExhibitDTO
		if err := json.Unmarshal(data, &exhibits); err != nil {
			http.Error(w, "Error parsing exhibits.json", http.StatusInternalServerError)
			log.Printf("Error unmarshalling json: %v", err)
			return
		}

		// Filter by IDs if provided
		idsParam := r.URL.Query().Get("ids")
		var filteredExhibits []ExhibitDTO

		if idsParam != "" {
			// Parse IDs from comma-separated string
			var targetIDs []int
			for _, idStr := range strings.Split(idsParam, ",") {
				id, err := strconv.Atoi(strings.TrimSpace(idStr))
				if err == nil {
					targetIDs = append(targetIDs, id)
				}
			}

			// Filter exhibits
			for _, exhibit := range exhibits {
				for _, targetID := range targetIDs {
					if exhibit.ID == targetID {
						filteredExhibits = append(filteredExhibits, exhibit)
						break
					}
				}
			}
		} else {
			filteredExhibits = exhibits
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(filteredExhibits); err != nil {
			log.Printf("Error encoding response: %v", err)
		}
	})

	http.HandleFunc("/artefacts", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		filePath, err := findFile("qm_data.json")
		if err != nil {
			http.Error(w, "Error finding `qm_data.json`: " +err.Error(), http.StatusInternalServerError)
			log.Printf("Error finding file: %v", err)
			return
		}
		data, err := os.ReadFile(filePath)
		if err != nil {
			http.Error(w, "Error reading qm_data.json: " +err.Error(), http.StatusInternalServerError)
			log.Printf("Error reading file %s: %v", filePath, err)
			return
		}

		var qmResponse QMResponse
		if err := json.Unmarshal(data, &qmResponse); err != nil {
			http.Error(w, "Error parsing qm_data.json: "  +err.Error(), http.StatusInternalServerError)
			log.Printf("Error unmarshalling json: %v", err)
			return
		}

		// Filter by objectNumbers if provided
		objectNumbersParam := r.URL.Query().Get("objectNumbers")
		var filteredArtefacts []ArtefactDTO

		if objectNumbersParam != "" {
			// Parse objectNumbers from comma-separated string
			var targetObjectNumbers []string
			for _, objNum := range strings.Split(objectNumbersParam, ",") {
				targetObjectNumbers = append(targetObjectNumbers, strings.TrimSpace(objNum))
			}

			// Filter artefacts
			for _, artefact := range qmResponse.Results {
				for _, targetObjNum := range targetObjectNumbers {
					if artefact.ObjectNumber == targetObjNum {
						filteredArtefacts = append(filteredArtefacts, artefact)
						break
					}
				}
			}
		} else {
			filteredArtefacts = qmResponse.Results
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(filteredArtefacts); err != nil {
			log.Printf("Error encoding response: %v", err)
		}
	})

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
