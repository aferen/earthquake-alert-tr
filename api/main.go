package main

import (
	"bufio"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/umahmood/haversine"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"bytes"
)

type Earthquake struct {
	Date             time.Time
	Coordinate       haversine.Coord
	Depth            float64
	MagnitudeMD      float64
	MagnitudeML      float64
	MagnitudeMW      float64
	Region           string
	DistancetoOrigin float64
}

var (
	earthquakes        []Earthquake
	earthquake         Earthquake
	line               string
	date               time.Time
	coordinate         haversine.Coord
	depth              float64
	magnitudeMD        float64
	magnitudeML        float64
	magnitudeMW        float64
	distanceToOrigin   float64
	region             string
	latitude           float64
	longitude          float64
	title              string
	message            string
)

func main() {
	listenAddr := os.Getenv("LISTEN_ADDR")
	addr := listenAddr + `:` + os.Getenv("PORT")
	http.HandleFunc("/send", sendNotification)
	log.Printf("starting server at %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func sendNotification(w http.ResponseWriter, r *http.Request) {
	fcmUrl := "https://fcm.googleapis.com/fcm/send" 
	origin := haversine.Coord{Lat: 40.770727, Lon: 29.118538}
	//origin := haversine.Coord{Lat: 37.444156, Lon: 37.188217}
	maxDistance := 100.0
	dateLayout := "2006.01.02T15:04:05 -0700"
	timezone := "Europe/Istanbul"
	maxTimeRange := 30.0 //minutes
	location, err := time.LoadLocation(timezone)
	now := time.Now().In(location)
	sendMessage := false
	counter := 0
	earthquakes = nil
	if err != nil {
		fmt.Println(err)
	}

	resp, err := http.Get("http://www.koeri.boun.edu.tr/scripts/lst0.asp")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 || err != nil {
		body, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		data := strings.Trim(body.Find("pre").First().Text(), "\n")
		scanner := bufio.NewScanner(strings.NewReader(data))
		for scanner.Scan() {
			if counter > 5 {
				line = strings.TrimSpace(strings.Replace(scanner.Text(), " ", ",", -1))
				for strings.Contains(line, ",,") {
					line = strings.TrimSpace(strings.Replace(line, ",,", ",", -1))
				}
				arr := strings.Split(line, ",")
				date, _ = time.Parse(dateLayout, strings.TrimSpace(arr[0])+"T"+strings.TrimSpace(arr[1])+" +0300")
				latitude, _ = strconv.ParseFloat(strings.TrimSpace(arr[2]), 64)
				longitude, _ = strconv.ParseFloat(strings.TrimSpace(arr[3]), 64)
				coordinate = haversine.Coord{Lat: latitude, Lon: longitude}
				depth, _ = strconv.ParseFloat(strings.TrimSpace(arr[4]), 64)
				magnitudeMD, _ = strconv.ParseFloat(strings.TrimSpace(arr[5]), 64)
				magnitudeML, _ = strconv.ParseFloat(strings.TrimSpace(arr[6]), 64)
				magnitudeMW, _ = strconv.ParseFloat(strings.TrimSpace(arr[7]), 64)
				region = strings.TrimSpace(arr[8]) + " " + strings.TrimSpace(arr[9])
				_, distanceToOrigin := haversine.Distance(origin, coordinate)
				earthquake = Earthquake{
					date,
					coordinate,
					depth,
					magnitudeMD,
					magnitudeML,
					magnitudeMW,
					region,
					distanceToOrigin,
				}
				if (earthquake.DistancetoOrigin < maxDistance && earthquake.MagnitudeML > 2.0) || earthquake.MagnitudeML >= 5.0 {
					if now.Sub(earthquake.Date).Minutes() < maxTimeRange {
						sendMessage = true
					}
					fmt.Println(earthquake.Region)
					earthquakes = append(earthquakes, earthquake)
				}
			}
			counter++
		}
	} else {
		log.Fatalf("failed to fetch data: %d %s", resp.StatusCode, resp.Status)
	}
	if sendMessage {
		fmt.Println("Notification is sending")
		title = fmt.Sprintf("Earthquake Happened!!!")
		message = ""
		
		for _, earthquake = range earthquakes {
			message = fmt.Sprintf("%s - %.1fML | %s | %s | Distance: %dkm\n", message, earthquake.MagnitudeML, earthquake.Date.Format("02/01/2006 15:04:05"), earthquake.Region, int(earthquake.DistancetoOrigin))
		}

		message = fmt.Sprintf("%s \n Total: %d", message, len(earthquakes))
		requestData := fmt.Sprintf("{\"to\": \"/topics/earthquake\",\"notification\": {\"title\": \"%s\",\"body\": \"%s\"}}", title, message)
		var jsonStr = []byte(requestData)
		req, err := http.NewRequest("POST", fcmUrl, bytes.NewBuffer(jsonStr))
		req.Header.Set("Authorization", os.Getenv("TOKEN"))
		req.Header.Set("Content-Type", "application/json")
	
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
	
		fmt.Println("Response Status:", resp.Status)
	}
}
