package twse_test

import (
	"fmt"
	"time"

	"github.com/chehsunliu/tshakutshai/pkg/client/twse"
)

func Example_basic() {
	var client = twse.NewClient(time.Second * 3)

	date := time.Date(2021, time.March, 29, 0, 0, 0, 0, time.UTC)

	dayQuotes, err := client.FetchDayQuotes(date)
	if err != nil {
		panic(fmt.Sprintf("error fetching day quotes: %v", err))
	}

	quotes, err := client.FetchDailyQuotes("2330", 2021, time.February)
	if err != nil {
		panic(fmt.Sprintf("error fetching daily quotes of TSMC: %v", err))
	}

	fmt.Printf("%s: %v\n", dayQuotes["2330"].Name, dayQuotes["2330"].Close)
	fmt.Printf("%s: %v\n", dayQuotes["2454"].Name, dayQuotes["2454"].Close)

	for i := range quotes {
		fmt.Printf("%v: %v\n", quotes[i].Date.Format("20060102"), quotes[i].Close)
	}
}
