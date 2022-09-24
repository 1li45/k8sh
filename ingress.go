package main

// Copyright 2022 Ilias Yacoubi (hi@ilias.sh)

// Goal of this application is to learn and get familiar with the client-go package.

import (
	"context"
	"flag"
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

func inspectIngress(i []v1.Ingress) ([]string, []string, []string, []string) {

	// slice for hosts
	var hs []string
	// slice for backend
	var bs []string
	// slice for annotation keys
	var ls []string
	// slice for whitelist
	var wl []string
	// slice for helm annotation
	var hl []string

	//var nameslice []*string
	//var namespaceslice []*string

	for value := range i {

		ingRuleHost := &i[value].Spec.Rules[0].Host
		ingRulePath := &i[value].Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Path
		ingBackendService := &i[value].Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Backend.ServiceName
		ingAnnotation := &i[value].Annotations

		// use regexp to replace these characters with nothing
		re, err := regexp.Compile(`[().*?$+]`)
		if err != nil {
			log.Fatal(err)
		}

		*ingRulePath = re.ReplaceAllString(*ingRulePath, "")

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

		for i, _ := range *ingAnnotation {
			ls = append(ls, i)

		}

		for _, j := range ls {
			// Check if nginx whitelist annotation is there.
			if j == "nginx.ingress.kubernetes.io/whitelist-source-range" {

				wl = append(wl, "🟢") //possible whitelist
			} else {
				wl = append(wl, "🔴") //no nginx whitelist
			}
			// Check if helm annotation is there.
			if j == "meta.helm.sh/release-name" {

				hl = append(hl, "🟢") //possible helm chart
			} else {
				hl = append(hl, "🔴") //no helm chart
			}

		}

	}

	return hs, bs, wl, hl

}

func statusChecker(s string) bool {
	_, err := http.Get(s)
	var resp bool

	if err != nil {
		resp = false
	} else {
		resp = true

	}
	return resp

}
