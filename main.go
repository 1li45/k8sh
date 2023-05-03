package main

// Copyright 2022 Ilias Yacoubi (hi@ilias.sh)

// Goal of this application is to learn and get familiar with the client-go package.

import (
	"flag"
	"fmt"
	"log"
)

var (
	ing bool
	pv  bool
)

func init() {
	flag.BoolVar(&ing, "ing", false, "Check cluster for dead Ingresses.")
	flag.BoolVar(&pv, "pv", false, "Check cluster for dead Persistant Volumes")
}

func main() {

	flag.Parse()
	var ans string
	clientset, err := getCluster()

	if ing {

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

	if pv {
		err = listStorageObjects(*&clientset)
		if err != nil {
			log.Fatalf("Could not list storage objects: %v", err)
		}

	}

}
