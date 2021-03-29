package twse

import (
	"fmt"
	"time"
)

func ExampleClient_FetchDayQuotes() {
	date := time.Date(2021, time.March, 30, 0, 0, 0, 0, time.UTC)

	client := NewClient()
	quotes, err := client.FetchDayQuotes(date)
	if err != nil {
		fmt.Printf("error fetching day quotes: %v", err)
		return
	}

	fmt.Printf("%s: %v\n", quotes["2330"].Name, quotes["2330"].Close)
}

func ExampleClient_FetchDailyQuotes() {
	client := NewClient()
	quotes, err := client.FetchDailyQuotes("2330", 2021, time.February)
	if err != nil {
		fmt.Printf("error fetching daily quotes of TSMC: %v", err)
		return
	}

	for i := range quotes {
		fmt.Printf("%v: %v\n", quotes[i].Date, quotes[i].Close)
	}
}
