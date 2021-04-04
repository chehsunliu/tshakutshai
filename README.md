# tshá-kú-tshài

[![ci](https://github.com/chehsunliu/tshakutshai/actions/workflows/ci.yml/badge.svg)](https://github.com/chehsunliu/tshakutshai/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/chehsunliu/tshakutshai/branch/main/graph/badge.svg?token=9BDEMWEZPQ)](https://codecov.io/gh/chehsunliu/tshakutshai)
[![Go Reference](https://pkg.go.dev/badge/github.com/chehsunliu/tshakutshai.svg)](https://pkg.go.dev/github.com/chehsunliu/tshakutshai)

This library provides simple HTTP clients to get basic information from the Taiwan Stock Exchange (TWSE) and Taipei Exchange (TPEx).

## Examples

Fetch the daily quotes of TSMC on March, 2021:

```go
package main

import (
	"fmt"
	"time"

	"github.com/chehsunliu/tshakutshai/pkg/client/twse"
)

func main() {
	// The interval between each query is set to 2 seconds. The TWSE
	// server is unlikely to ban you at this query frequency.
	var client = twse.NewClient(time.Second * 2)
	quotes, _ := client.FetchDailyQuotes("2330", 2021, time.March)
	
	for _, quote := range quotes {
		fmt.Printf("%v\n", quote)
    }
}
```

Fetch the monthly quotes of PChome in 2020:

```go
package main

import (
	"fmt"
	"time"

	"github.com/chehsunliu/tshakutshai/pkg/client/tpex"
)

func main() {
	// It seems the TPEx server never bans crawlers, so no need to
	// sleep between each query.
	var client = tpex.NewClient(0)
	quotes, _ := client.FetchMonthlyQuotes("8044", 2020)
	
	for _, quote := range quotes {
		fmt.Printf("%v\n", quote)
    }
}
```

You can also use your own HTTP clients:

```go
import (
	"net/http"
	
	"github.com/chehsunliu/tshakutshai/pkg/client/twse"
	"github.com/chehsunliu/tshakutshai/pkg/client/tpex"
)

var (
	twseClient = &twse.Client{HttpClient: &http.Client{}}
	tpexClient = &tpex.Client{HttpClient: &http.Client{}}
)
```

Please refer to [the online document](https://pkg.go.dev/github.com/chehsunliu/tshakutshai) for more details.
