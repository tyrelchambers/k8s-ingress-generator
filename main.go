package main

import (
	"fmt"
	"k8s/k8s"
	"net/http"
	"os"
)

func main() {

	localConfig := os.Getenv("LOCAL") == "true"

	k8sConfig := k8s.K8sClientStruct{
		Namespace: "dynamic-sites",
	}
	k8sConfig.Config(localConfig)
	k8s.K8sClient = &k8sConfig

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			k8s.Post(w, r)
		} else if r.Method == http.MethodDelete {
			k8s.Delete(w, r)
		}
	})

	fmt.Println("Listening on port 8080")
	http.ListenAndServe(":8080", mux)

}
