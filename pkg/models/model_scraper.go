package models

import (
	"encoding/json"
	"sync"
)

var scrapers []Scraper

type ScraperFunc func(*sync.WaitGroup, bool, []string, chan<- ScrapedScene) error

type Scraper struct {
	ID        string
	Name      string
	AvatarURL string
	Scrape    ScraperFunc
}

type ScrapedScene struct {
	SceneID     string   `json:"_id"`
	SiteID      string   `json:"scene_id"`
	SceneType   string   `json:"scene_type"`
	Title       string   `json:"title"`
	Studio      string   `json:"studio"`
	Site        string   `json:"site"`
	Covers      []string `json:"covers"`
	Gallery     []string `json:"gallery"`
	Tags        []string `json:"tags"`
	Cast        []string `json:"cast"`
	Filenames   []string `json:"filename"`
	Duration    int      `json:"duration"`
	Synopsis    string   `json:"synopsis"`
	Released    string   `json:"released"`
	HomepageURL string   `json:"homepage_url"`
	MembersUrl  string   `json:"members_url"`
	TrailerType string   `json:"trailer_type"`
	TrailerSrc  string   `json:"trailer_source"`
}

type TrailerScrape struct {
	SceneUrl       string `json:"scene_url"`        // url of the page to be scrapped
	HtmlElement    string `json:"html_element"`     // path to section of html (using colly)
	ExtractRegex   string `json:"extract_regex"`    // regex expression to extract the json, eg from a json variable assignment in javascript
	ContentBaseUrl string `json:"content_base_url"` // prefix for the url if the scrapped content urls are not abosolute
	RecordPath     string `json:"record_path"`      // points to a json array of video source (optional, there maybe a single video), uses jsonpath syntax
	ContentPath    string `json:"content_path"`     // points to the content url uses jsonpath syntax
	EncodingPath   string `json:"encoding_path"`    // optional, points to the encoding for the source using jsonpath syntax, eg h264, h265
	QualityPath    string `json:"quality_path"`     // points to the quality using jsonpath syntax, eg 1440p, 5k
}

func (s *ScrapedScene) ToJSON() ([]byte, error) {
	return json.Marshal(s)
}

func (s *ScrapedScene) Log() error {
	j, err := json.MarshalIndent(s, "", "  ")
	log.Debugf("\n%v", string(j))
	return err
}

func GetScrapers() []Scraper {
	return scrapers
}

func RegisterScraper(id string, name string, avatarURL string, f ScraperFunc) {
	s := Scraper{}
	s.ID = id
	s.Name = name
	s.AvatarURL = avatarURL
	s.Scrape = f
	scrapers = append(scrapers, s)
}
