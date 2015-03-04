package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/blevesearch/bleve"
)

type Event struct {
	UID         string    `json:"uid"`
	Summary     string    `json:"summary"`
	Description string    `json:"description"`
	Speaker     string    `json:"speaker"`
	Start       time.Time `json:"start"`
	Duration    float64   `json:"duration"`
}

func main() {
	mapping := bleve.NewIndexMapping()
	index, err := bleve.New("gopherconin.bleve", mapping)
	if err != nil {
		log.Fatal(err)
	}

	count := 0
	batch := bleve.NewBatch() // HL
	for event := range parseEvents() {
		batch.Index(event.UID, event) // HL
		if batch.Size() > 100 {
			err := index.Batch(batch) // HL
			if err != nil {
				log.Fatal(err)
			}
			count += batch.Size()
			batch = bleve.NewBatch()
		}
	}
	if batch.Size() > 0 {
		index.Batch(batch)
		if err != nil {
			log.Fatal(err)
		}
		count += batch.Size()
	}
	fmt.Printf("Indexed %d Events\n", count)
}

const timeFormat = "2006-01-02T15:04"

// just a quick and dirty way to parse the schedule
func parseEvents() chan *Event {
	rv := make(chan *Event)

	go func() {
		defer close(rv)

		file, err := os.Open("gophercon-schedule.html")
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		doc, err := goquery.NewDocumentFromReader(file)
		if err != nil {
			log.Fatal(err)
		}

		loc, err := time.LoadLocation("Asia/Kolkata")
		if err != nil {
			log.Fatal(err)
		}

		doc.Find("div.panel").Each(func(i int, s0 *goquery.Selection) {
			date := ""
			s0.Find("div.panel-heading").Each(func(i int, s1 *goquery.Selection) {
				s1.Find("sup").Remove()
				date = s1.Text()
			})

			s0.Find("table tr").Each(func(i int, s2 *goquery.Selection) {
				event := Event{}
				s2.Find("td").Each(func(i int, s3 *goquery.Selection) {
					if i == 0 {
						//eventDate = date + " " + s3.Text()
					} else if i == 1 {
						s3.Find("h2.modal-title").Each(func(i int, s4 *goquery.Selection) {
							if i == 0 {
								event.Summary = s4.Text()
								event.UID = strings.ToLower(event.Summary)
								event.UID = strings.Replace(event.UID, " ", "_", -1)
							}
						})
						s3.Find("p").Each(func(i int, s4 *goquery.Selection) {
							newText := s4.Text()
							newText = strings.Replace(newText, "\t", "", -1)
							newText = strings.Replace(newText, "\n", "", -1)
							event.Description += newText + " "
						})
						s3.Find("div.modal-desc").Each(func(i int, s5 *goquery.Selection) {
							if i == 0 {
								eventDateStr := strings.TrimSpace(s5.Text())
								if strings.HasSuffix(eventDateStr, " (IST)") {
									eventDateStr = eventDateStr[0 : len(eventDateStr)-6]
								}
								formattedDate := "2015-02-21T"
								if strings.Contains(eventDateStr, "20th") {
									formattedDate = "2015-02-20T"
								}
								ypos := strings.Index(eventDateStr, "2015")
								if ypos > 0 {
									rest := eventDateStr[ypos+4:]
									topos := strings.Index(rest, "to")
									if topos > 0 {
										start := rest[0:topos]
										start = strings.TrimSpace(start)
										end := rest[topos+3:]
										end = strings.TrimSpace(end)
										startTime, err := parseWeirdTime(start)
										if err == nil {
											parsedStart, err := time.ParseInLocation(timeFormat, formattedDate+startTime, loc)
											if err != nil {
												log.Fatalf("error parsing start time: %v", err)
											} else {
												event.Start = parsedStart
											}
											endTime, err := parseWeirdTime(end)
											if err == nil {
												parsedEnd, err := time.ParseInLocation(timeFormat, formattedDate+endTime, loc)
												if err != nil {
													log.Fatalf("error parsing end time: %v", err)
												} else {
													duration := parsedEnd.Sub(parsedStart).Minutes()
													event.Duration = duration
												}
											}
										}
									}

								}
							}
						})
						s3.Find("span.name").Each(func(i int, s5 *goquery.Selection) {
							if i == 0 {
								event.Speaker = strings.TrimSpace(s5.Text())
							}
						})

					}
				})
				if event.Summary != "" {
					rv <- &event
				}
			})
		})

	}()

	return rv
}

func parseWeirdTime(t string) (string, error) {
	pm := false
	if strings.HasSuffix(t, " pm") {
		pm = true
		t = t[0 : len(t)-3]
	} else if strings.HasSuffix(t, " am") {
		t = t[0 : len(t)-3]
	} else if strings.HasSuffix(t, " noon") {
		pm = true
		t = t[0 : len(t)-5]
	}

	colonPos := strings.Index(t, ":")
	if colonPos < 0 {
		// try dot
		colonPos = strings.Index(t, ".")
		if colonPos < 1 {
			return "", fmt.Errorf("colon was: %d, time was %s", colonPos, t)
		}
	}
	hourStr := t[0:colonPos]
	minuteStr := t[colonPos+1:]
	hour, err := strconv.Atoi(hourStr)
	if err != nil {
		return "", fmt.Errorf("err: %v for %s", err, t)
	}
	if pm && hour != 12 {
		hour += 12
	}
	minute, err := strconv.Atoi(minuteStr)
	if err != nil {
		return "", fmt.Errorf("err: %v for %s", err, t)
	}
	return fmt.Sprintf("%d:%02d", hour, minute), nil
}

func init() {
	os.RemoveAll("gopherconin.bleve")
}
