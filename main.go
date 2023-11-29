package main

import (
	"context"
	"fmt"
	"log"

	"github.com/digilolnet/bunnynetedgeips/pkg/bunnynetedgeips"
)

func main() {
	ips, err := bunnynetedgeips.BunnynetEdgeIPs(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
	for _, ip := range ips {
		fmt.Println(ip)
	}
}
