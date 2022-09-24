package main

// Copyright 2022 Ilias Yacoubi (hi@ilias.sh)

// Goal of this application is to learn and get familiar with the client-go package.

import (
	"flag"
	"fmt"
)

var (
	ing bool
)

func init() {
	flag.BoolVar(&ing, "ing", false, "Check cluster for dead Ingresses.")
}

func main() {

	flag.Parse()

	if ing {
		clientset, _ := getCluster()
		ingItems, _ := getIngress(*clientset)
		hs, _, wl, hl := inspectIngress(ingItems)
		i := 0
		for _, host := range hs {
			url := "http://" + host

			if !statusChecker(url) {
				fmt.Printf("🔴 %s \n\t Whitelist: %s \n\t Helm: %s\n", host, wl[i], hl[i])

			}

			i++

		}
		fmt.Printf("\n🔎💻 %d URL's \n", len(hs))

	}

}
