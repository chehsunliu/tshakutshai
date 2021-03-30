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

	// Output:
	// 台積電: 599
	// 聯發科: 941
	// 20210201: 611
	// 20210202: 632
	// 20210203: 630
	// 20210204: 627
	// 20210205: 632
	// 20210217: 663
	// 20210218: 660
	// 20210219: 652
	// 20210222: 650
	// 20210223: 641
	// 20210224: 625
	// 20210225: 635
	// 20210226: 606
}
