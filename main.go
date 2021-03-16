package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"
)

var loginPage = "https://hudsoncovidvax.org/login"
var dashboardPage = "https://hudsoncovidvax.org/member/list"

// used to find the user specific appointments page link
var appointmentsPrefix = "https://hudsoncovidvax.org/second/appt"

// text present when no appointments are available
var noAptsRegex = regexp.MustCompile(`NOT ABLE TO SCHEDULE`)

var userAgent = `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.82 Safari/537.36`

func main() {

	email := flag.String("email", "", "email address")
	password := flag.String("pass", "", "password")
	flag.Parse()

	if *email == "" || *password == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	tokenCollector := colly.NewCollector(colly.UserAgent(userAgent))

	tokenCollector.OnHTML("html", func(e *colly.HTMLElement) {
		// grab the CSRF token
		csrfToken := strings.Split(e.ChildAttr("meta[name=\"csrf-token\"]", "content"), "\n")[0]

		if csrfToken == "" {
			log.Fatal("Unable to scrape CSRF token")
		}

		loginCollector := tokenCollector.Clone()

		// use CSRF token scraped from the previous request to login
		err := loginCollector.Post(loginPage, map[string]string{"_token": csrfToken, "email": *email, "password": *password})
		if err != nil {
			log.Fatal(err)
		}

		// find link to appointments
		loginCollector.OnHTML("a", func(e *colly.HTMLElement) {
			link := e.Attr("href")
			if strings.HasPrefix(link, appointmentsPrefix) {

				appointmentCollector := loginCollector.Clone()

				appointmentCollector.OnHTML("body", func(e *colly.HTMLElement) {
					if noAptsRegex.MatchString(e.Text) {
						fmt.Println("No appointments")
						os.Exit(1)
					} else {
						fmt.Println("Appointments available?")
						os.Exit(0)
					}
				})

				// 3) visit appointments link
				appointmentCollector.Visit(link)
			}
		})

		// 2) visit dashboard
		loginCollector.Visit(dashboardPage)

	})

	// 1) visit login page
	tokenCollector.Visit(loginPage)
}
