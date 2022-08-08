package main

import (
	"os"
	"scraper/scraper"
	"strings"

	"github.com/labstack/echo"
)

const FileName string = "jobs.csv"

func handleHome(c echo.Context) error {
	return c.File("home.html")
}
func handleScrape(c echo.Context) error {
	defer os.Remove(FileName)
	term := strings.ToLower(scraper.CleanString(c.FormValue("term")))
	scraper.Scrape(term)
	// 첨부파일 리턴한는 함수
	return c.Attachment(FileName, FileName)
}
func main() {
	scraper.Scrape("term")
	e := echo.New()
	e.GET("/", handleHome)
	e.POST("/scrape", handleScrape)
	e.Logger.Fatal(e.Start(":1323"))
}
