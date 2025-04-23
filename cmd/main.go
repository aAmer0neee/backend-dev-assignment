package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/aAmer0neee/backend-dev-assignment/pkg/exchange"
)

var (
	periodFlag = flag.Int("period", 1, "specify days amount")
)

func main() {
	timer := time.Now()
	flag.Parse()

	info := exchange.New(*periodFlag)
	info.Exchange()

	fmt.Println("время работы", time.Since(timer))
}
