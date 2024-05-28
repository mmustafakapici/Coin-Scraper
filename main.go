package main

import (
	"encoding/csv"
	"fmt"
	"github.com/gocolly/colly"
	"log"
	"os"
	_ "strconv"
	"sync"
)

type Coin struct {
	Name   string
	Price  string
	Change string
}

func main() {
	coins := collectCoinData()
	writeCSV(coins)
}

func collectCoinData() []Coin {
	var wg sync.WaitGroup
	coinChan := make(chan Coin, 100)
	coins := []Coin{}

	go func() {
		for coin := range coinChan {
			coins = append(coins, coin)
		}
	}()

	wg.Add(1)
	go fetchCoinData(&wg, coinChan)

	wg.Wait()
	close(coinChan)

	return coins
}

func fetchCoinData(wg *sync.WaitGroup, coinChan chan<- Coin) {
	defer wg.Done()

	c := colly.NewCollector()

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting:", r.URL)
	})
	c.OnError(func(_ *colly.Response, err error) {
		log.Println("Something went wrong:", err)
	})

	c.OnHTML("tbody tr", func(e *colly.HTMLElement) {
		coin := Coin{
			Name:   e.ChildText("p[class='sc-1eb5slv-0 iJjGCS']"),
			Price:  e.ChildText("div[class='sc-131di3y-0 cLgOOr']"),
			Change: e.ChildText("span[class='sc-15yy2pl-0 kAXKAX']"),
		}

		if coin.Name != "" && coin.Price != "" && coin.Change != "" {
			coinChan <- coin
		} else {
			log.Println("Failed to extract data for a coin")
		}
	})

	c.Visit("https://coinmarketcap.com/")
}

func writeCSV(coins []Coin) {
	file, err := os.Create("coins.csv")
	if err != nil {
		log.Fatalln("Failed to create output CSV file:", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{"Name", "Price", "Change"}
	if err := writer.Write(headers); err != nil {
		log.Fatalln("Failed to write headers to CSV file:", err)
	}

	for i, coin := range coins {
		if i >= 100 {
			break
		}
		record := []string{coin.Name, coin.Price, coin.Change}
		if err := writer.Write(record); err != nil {
			log.Fatalln("Failed to write record to CSV file:", err)
		}
	}

	fmt.Println("Finished collecting and writing coin data.")
}
