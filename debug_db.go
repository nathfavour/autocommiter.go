package main

import (
	"github.com/nathfavour/autocommiter.go/internal/index"
	"fmt"
)

func main() {
	fmt.Println("Full Cache Listing:")
	index.ListAllCache()
}
