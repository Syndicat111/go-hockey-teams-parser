package parser

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

var wg sync.WaitGroup

type Team struct {
	Name       string  `json:"name"`
	Year       int     `json:"year"`
	Wins       int     `json:"wins"`
	Losses     int     `json:"losses"`
	OTLosses   int     `json:"ot_losses,omitempty"`
	WinPercent float64 `json:"win_percent"`
	GF         int     `json:"goals_for"`
	GA         int     `json:"goals_against"`
	Diff       int     `json:"diff"`
}

const (
	baseUrl    = "https://www.scrapethissite.com/pages/forms/?page_num=%d"
	totalPages = 24
	perPage    = 25
)

func fetch(pageNum int) ([]byte, error) {
	resp, err := http.Get(fmt.Sprintf(baseUrl, pageNum))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("request failed for page %d: %s", pageNum, resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		return nil, err
	}
	resp.Body.Close()
	return body, nil
}

func parse(pageNum int, c chan<- Team) {
	defer wg.Done()
	body, err := fetch(pageNum)
	if err != nil {
		log.Println(err)
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		log.Println(err)
	}
	doc.Find(".team").Each(func(i int, s *goquery.Selection) {
		year, _ := strconv.Atoi(strings.TrimSpace(s.Find(".year").Text()))
		wins, _ := strconv.Atoi(strings.TrimSpace(s.Find(".wins").Text()))
		losses, _ := strconv.Atoi(strings.TrimSpace(s.Find(".losses").Text()))
		otLosses, err := strconv.Atoi(strings.TrimSpace(s.Find(".ot-losses").Text()))
		if err != nil {
			otLosses = 0
		}
		winPercent, _ := strconv.ParseFloat(strings.TrimSpace(s.Find(".pct").Text()), 64)
		gf, _ := strconv.Atoi(strings.TrimSpace(s.Find(".gf").Text()))
		ga, _ := strconv.Atoi(strings.TrimSpace(s.Find(".ga").Text()))
		diff, _ := strconv.Atoi(strings.TrimSpace(s.Find(".diff").Text()))
		team := &Team{
			Name:       strings.TrimSpace(s.Find(".name").Text()),
			Year:       year,
			Wins:       wins,
			Losses:     losses,
			OTLosses:   otLosses,
			WinPercent: winPercent,
			GF:         gf,
			GA:         ga,
			Diff:       diff,
		}
		c <- *team
	})
}

func CollectTeams() []Team {
	teams := make(chan Team, totalPages*perPage)
	teamsSlice := make([]Team, 0, totalPages*perPage)

	for i := 1; i <= totalPages; i++ {
		wg.Add(1)
		go parse(i, teams)
	}
	wg.Wait()
	close(teams)
	for team := range teams {
		teamsSlice = append(teamsSlice, team)
	}

	return teamsSlice
}
