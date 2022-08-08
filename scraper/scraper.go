package scraper

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type extractedJob struct {
	title    string
	location string
	salary   string
	summary  string
}

//scrapper
func Scrape(term string) {
	var baseURL string = "https://kr.indeed.com/jobs?q=" + term + "&limit=50"

	var jobs []extractedJob
	c := make(chan []extractedJob)
	startTime := time.Now()
	totalPage := getPages(baseURL)

	for i := 0; i < totalPage; i++ {
		go getPage(i, baseURL, c)
	}
	for i := 0; i < totalPage; i++ {
		extractedJob := <-c
		jobs = append(jobs, extractedJob...)

	}
	writeJobs(jobs)
	fmt.Println("Done, extract", len(jobs), "jobs from indeed.com")
	endTime := time.Now()
	fmt.Println("Operating time: ", endTime.Sub(startTime))

}

func getPage(page int, url string, mainC chan<- []extractedJob) {
	var jobs []extractedJob
	c := make(chan extractedJob)

	pageURL := url + "&start=" + strconv.Itoa(page*10)

	fmt.Println(pageURL)

	res, err := http.Get(pageURL)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()
	// 처음 고쿼리 사용
	doc, err := goquery.NewDocumentFromReader(res.Body)

	checkErr(err)

	searchCards := doc.Find(".job_seen_beacon")
	searchCards.Each(func(i int, card *goquery.Selection) {
		go extractJob(card, c)
		//jobs = append(jobs, job)

	})

	for i := 0; i < searchCards.Length(); i++ {
		job := <-c
		jobs = append(jobs, job)

	}
	mainC <- jobs
}

func extractJob(card *goquery.Selection, c chan<- extractedJob) {
	title := CleanString(card.Find(".jobTitle > a").Text())
	location := CleanString(card.Find(".companyLocation").Text())
	salary := CleanString(card.Find(".attribute_snippet").Text())
	summary := CleanString(card.Find(".job-snippet").Text())
	c <- extractedJob{
		title:    title,
		location: location,
		salary:   salary,
		summary:  summary}
}

func CleanString(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}

func getPages(url string) int {
	pages := 0
	res, err := http.Get(url)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	doc.Find(".pagination").Each(func(i int, s *goquery.Selection) {
		pages = s.Find("a").Length()
	})
	return pages
}

func writeJobs(jobs []extractedJob) {
	file, err := os.Create("jobs.csv")
	checkErr(err)
	c := make(chan []string)
	w := csv.NewWriter(file)
	// defer  함수가끝날때 실행되는것
	defer w.Flush()

	headers := []string{"Title", "Location", "Salary", "summary"}

	wErr := w.Write(headers)
	checkErr(wErr)

	for _, job := range jobs {
		go eachWritejob(job, c)

	}

	for i := 0; i < len(jobs); i++ {
		jwErr := w.Write(<-c)
		checkErr(jwErr)
	}
}

func eachWritejob(job extractedJob, c chan<- []string) {
	c <- []string{job.title, job.location, job.salary, job.summary}
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}

}

func checkCode(res *http.Response) {

	if res.StatusCode != 200 {
		log.Fatalln("Request failed with Status:", res.StatusCode)
	}

}
