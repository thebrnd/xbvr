package scrape

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly"
	"github.com/mozillazg/go-slugify"
	"github.com/thoas/go-funk"
	"github.com/xbapps/xbvr/pkg/models"
)

func VRPorn(wg *sync.WaitGroup, updateSite bool, knownScenes []string, out chan<- models.ScrapedScene, scraperID string, siteID string, company string) error {
	defer wg.Done()
	logScrapeStart(scraperID, siteID)

	sceneCollector := createCollector("vrporn.com")
	siteCollector := createCollector("vrporn.com")

	// RegEx Patterns
	sceneIDRegEx := regexp.MustCompile(`^post-(\d+)`)
	dateRegEx := regexp.MustCompile(`(?i)^VideoPosted (?:on Premium )?on (.+)$`)
	durationRegEx := regexp.MustCompile(`var timeAfter="(?:(\d+):)?(\d+):(\d+)";`)

	sceneCollector.OnHTML(`html`, func(e *colly.HTMLElement) {
		if !dateRegEx.MatchString(e.ChildText(`div.content-box.posted-by-box.posted-by-box-sub span.footer-titles`)) {
			// VRPorn hosts VR games, apparently
			return
		}

		sc := models.ScrapedScene{}
		sc.SceneType = "VR"
		sc.Studio = company
		sc.Site = siteID
		sc.HomepageURL = strings.Split(e.Request.URL.String(), "?")[0]

		// Scene ID - get from page HTML
		id := sceneIDRegEx.FindStringSubmatch(strings.TrimSpace(e.ChildAttr(`article.post`, "class")))[1]
		sc.SiteID = id
		sc.SceneID = slugify.Slugify(sc.Site) + "-" + sc.SiteID

		// Title
		e.ForEach(`h1.content-title`, func(id int, e *colly.HTMLElement) {
			sc.Title = strings.TrimSpace(e.Text)
		})

		// Cover
		coverURL := e.ChildAttr("#dl8videoplayer", "poster")
		if len(coverURL) > 0 {
			sc.Covers = append(sc.Covers, coverURL)
		}

		// Gallery
		e.ForEach(`.vrp-gallery-pro a`, func(id int, e *colly.HTMLElement) {
			sc.Gallery = append(sc.Gallery, e.Request.AbsoluteURL(e.Attr("href")))
		})

		// Synopsis
		e.ForEach(`.entry-content.post-video-description`, func(id int, e *colly.HTMLElement) {
			sc.Synopsis = strings.TrimSpace(e.Text)
		})

		// Skipping some very generic and useless tags
		skiptags := map[string]bool{
			"3D":     true,
			"60 FPS": true,
			"HD":     true,
		}

		// Tags
		e.ForEach(`.tag-box a[rel="tag"]`, func(id int, e *colly.HTMLElement) {
			trimmed := strings.TrimSpace(e.Text)
			if !skiptags[trimmed] {
				sc.Tags = append(sc.Tags, trimmed)
			}
		})

		// Cast
		e.ForEach(`.name_pornstar`, func(id int, e *colly.HTMLElement) {
			sc.Cast = append(sc.Cast, strings.TrimSpace(e.Text))
		})

		// trailer details
		sc.TrailerType = "scrape_html"
		params := models.TrailerScrape{SceneUrl: sc.HomepageURL, HtmlElement: "dl8-video source", ContentPath: "src", QualityPath: "quality"}
		strParams, _ := json.Marshal(params)
		sc.TrailerSrc = string(strParams)

		// Release Date
		date := dateRegEx.FindStringSubmatch(e.ChildText(`div.content-box.posted-by-box.posted-by-box-sub span.footer-titles`))[1]
		if len(date) > 0 {
			dt, _ := time.Parse("January 02, 2006", date)
			sc.Released = dt.Format("2006-01-02")
		}

		// Duration
		e.ForEachWithBreak(`script`, func(id int, e *colly.HTMLElement) bool {
			var duration int
			m := durationRegEx.FindStringSubmatch(e.Text)
			if len(m) == 4 {
				hours, _ := strconv.Atoi("0" + m[1])
				minutes, _ := strconv.Atoi(m[2])
				duration = hours*60 + minutes
			}
			sc.Duration = duration
			return duration == 0
		})

		out <- sc
	})

	siteCollector.OnHTML(`div.pagination a.next`, func(e *colly.HTMLElement) {
		pageURL := e.Request.AbsoluteURL(e.Attr("href"))
		siteCollector.Visit(pageURL)
	})

	siteCollector.OnHTML(`body.tax-studio article.post div.tube-post a`, func(e *colly.HTMLElement) {
		sceneURL := e.Request.AbsoluteURL(e.Attr("href"))
		// If scene exists in database, there's no need to scrape
		if !funk.ContainsString(knownScenes, sceneURL) {
			sceneCollector.Visit(sceneURL)
		}
	})

	siteCollector.Visit("https://vrporn.com/studio/" + scraperID + "/?sort=newest")

	if updateSite {
		updateSiteLastUpdate(scraperID)
	}
	logScrapeFinished(scraperID, siteID)
	return nil
}

func addVRPornScraper(id string, name string, company string, avatarURL string) {
	registerScraper(id, name+" (VRPorn)", avatarURL, func(wg *sync.WaitGroup, updateSite bool, knownScenes []string, out chan<- models.ScrapedScene) error {
		return VRPorn(wg, updateSite, knownScenes, out, id, name, company)
	})
}

func init() {
	//addVRPornScraper("evileyevr", "EvilEyeVR", "EvilEyeVR", "https://mcdn.vrporn.com/files/20190605151715/evileyevr-logo.jpg")
	addVRPornScraper("randysroadstop", "Randy's Road Stop", "NaughtyAmerica", "https://mcdn.vrporn.com/files/20170718073527/randysroadstop-vr-porn-studio-vrporn.com-virtual-reality.png")
	addVRPornScraper("realteensvr", "Real Teens VR", "NaughtyAmerica", "https://mcdn.vrporn.com/files/20170718063811/realteensvr-vr-porn-studio-vrporn.com-virtual-reality.png")
	addVRPornScraper("vrclubz", "VRClubz", "VixenVR", "https://mcdn.vrporn.com/files/20200421094123/vrclubz_logo_NEW-400x400_webwhite.png")
}
