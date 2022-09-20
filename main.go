package main

// Copyright 2022 Ilias Yacoubi (hi@ilias.sh)

// Goal of this application is to learn and get familiar with the client-go package.
// This application checks for 'dead' ingresses in a cluster, backs it up and deletes it from the cluster.

// TODO:
// INGRESS FUNCTION:
// - Check if there are multiple Hosts and Paths per Item in getSlug function
// - Implement Goroutines.
// - Backup dead Ingress (in yaml) and delete it.
//
// NEW FEUTURES:
// - Check for dead Volums by checking read and write history.

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"

	v1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func getCluster() (*kubernetes.Clientset, error) {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// return clientset
	return kubernetes.NewForConfig(config)

}

// input clientset. return ingress items.
func getIngress(clientset kubernetes.Clientset) ([]v1.Ingress, error) {

	// get all ingresses
	ingresses, err := clientset.ExtensionsV1beta1().Ingresses("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	// return ingress items
	return ingresses.Items, nil

}

// get full slug from ingress items
func getSlug(i []v1.Ingress) ([]string, []string) {

	// create hostSlice and BackendSlice
	var hs []string
	var bs []string

	for value := range i {

		ingRuleHost := &i[value].Spec.Rules[0].Host
		ingRulePath := &i[value].Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Path
		ingBackendService := &i[value].Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Backend.ServiceName

		// use regexp to replace these characters with nothing
		re, err := regexp.Compile(`[().*?$+]`)
		if err != nil {
			log.Fatal(err)
		}

		*ingRulePath = re.ReplaceAllString(*ingRulePath, "")

		//fmt.Printf("%s \n", *ingRuleHost)

		// look for ' | ' in path, split it and put value in a slice.
		split := strings.Split(*ingRulePath, "|")
		for _, value := range split {
			// check is value doesn't start with '/', if not add '/'.
			if !strings.HasPrefix(value, "/") {
				value = "/" + value

			}
			fullSlug := *ingRuleHost + value
			hs = append(hs, fullSlug)
			bs = append(bs, *ingBackendService)

		}

	}
	return hs, bs

}

func statusChecker(s string) {
	_, err := http.Get(s)

	if err != nil {
		fmt.Printf("🔴 %s\n", s)
	} else {
		fmt.Printf("🟢 %s\n", s)
	}

}

func main() {
	clientset, _ := getCluster()
	ingItems, _ := getIngress(*clientset)
	hs, _ := getSlug(ingItems)

	for _, host := range hs {
		//fmt.Println(host)
		url := "http://" + host
		statusChecker(url)

	}

}
