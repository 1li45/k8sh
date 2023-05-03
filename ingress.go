package main

import (
	"context"
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"

	net "k8s.io/api/networking/v1"
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
func getIngress(clientset kubernetes.Clientset) ([]net.Ingress, error) {

	// get all ingresses
	ingresses, err := clientset.NetworkingV1().Ingresses("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	// return ingress items
	return ingresses.Items, nil

}

func inspectIngress(i []net.Ingress) ([]string, []string, []bool, []bool, []string, []string) {

	// slice for hosts
	var hs []string
	// slice for backend
	var bs []string
	// slice for annotation keys
	var ls []string
	// slice for whitelist
	var wl []bool
	// slice for helm annotation
	var hl []bool

	var in []string
	var ins []string

	for value := range i {

		ingName := &i[value].Name
		ingNamespace := &i[value].Namespace
		ingRuleHost := &i[value].Spec.Rules[0].Host
		ingRulePath := &i[value].Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Path
		ingBackendService := &i[value].Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Backend.Service.Name
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
			//bs = append(bs, *ingBackendService)

			for i, _ := range *ingAnnotation {
				ls = append(ls, i)

			}

			for _, j := range ls {
				// Check if nginx whitelist annotation is there.
				if j == "nginx.ingress.kubernetes.io/whitelist-source-range" {

					wl = append(wl, true) //possible whitelist
				} else {
					wl = append(wl, false) //no nginx whitelist
				}
				// Check if helm annotation is there.
				if j == "meta.helm.sh/release-name" || j == "helm.sh/chart" {

					hl = append(hl, true) //possible helm chart
				} else {
					hl = append(hl, false) //no helm chart
				}

			}

			bs = append(bs, *ingBackendService)
			in = append(in, *ingName)
			ins = append(ins, *ingNamespace)

		}

	}

	return hs, bs, wl, hl, in, ins

}

func statusChecker(s string) bool {
	// ignore expired certificates
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	_, err := http.Get(s)
	var resp bool

	if err != nil {
		resp = false
	} else {
		resp = true

	}
	return resp

}

func deleteIngress(clientset kubernetes.Clientset, name string, namespace string) {
	err := clientset.NetworkingV1().Ingresses(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		log.Fatal(err)
	}

}
