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
	var ans string

	if ing {
		clientset, _ := getCluster()
		ingItems, _ := getIngress(*clientset)
		hs, _, wl, hl, in, ins := inspectIngress(ingItems)
		i := 0

		for _, host := range hs {
			url := "http://" + host

			if !statusChecker(url) && !wl[i] && !hl[i] {
				fmt.Printf("🔴 %s \n", host)
				fmt.Printf("Delete ingress %s in %s y/n: ", in[i], ins[i])
				fmt.Scanln(&ans)
				if ans == "Y" || ans == "y" {
					deleteIngress(*clientset, in[i], ins[i])

				}

			}

			i++

		}
		fmt.Printf("\n🔎💻 %d URL's \n", len(hs))

	}

}
