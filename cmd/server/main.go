package main

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	MemStorage "github.com/ArtemShalinFe/metcoll/internal"
)

func counterHandler(w http.ResponseWriter, r *http.Request) {

	k, v := getKeyValueMetric(r.URL)
	if k == "" {
		http.Error(w, "name metric is empty", http.StatusNotFound)
		return
	}

	value, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newValue := value + int64(MemStorage.Values.Get(k))

	MemStorage.Values.Set(k, uint64(newValue))

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	fmt.Printf("%s:%v", k, newValue)

}

func gaugeHandler(w http.ResponseWriter, r *http.Request) {

	k, v := getKeyValueMetric(r.URL)
	if k == "" {
		http.Error(w, "name metric is empty", http.StatusNotFound)
		return
	}

	newValue, err := strconv.ParseFloat(v, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	MemStorage.Values.Set(k, uint64(newValue))

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	fmt.Printf("%s:%v", k, newValue)

}

func middleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			w.Header().Add("Allow", "POST")
			http.Error(w, fmt.Sprintf("The method %s is not allowed. The POST method is allowed.", r.Method), http.StatusMethodNotAllowed)
			return
		}

		if r.Header.Get("Content-Type") != "text/plain" {
			http.Error(w, "Want Content-Type: text/plain in headers", http.StatusUnsupportedMediaType)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getKeyValueMetric(uri *url.URL) (string, string) {

	key := ""
	value := ""

	re := regexp.MustCompile(`/update/(counter|gauge)/`)
	metric := re.ReplaceAllString(uri.RequestURI(), "")
	metrics := strings.Split(metric, "/")

	if len(metrics) == 2 {
		key = metrics[0]
		value = metrics[1]
	}

	return key, value

}

func main() {

	mux := http.NewServeMux()
	mux.Handle(`/update/counter/`, middleware(http.HandlerFunc(counterHandler)))
	mux.Handle(`/update/gauge/`, middleware(http.HandlerFunc(gaugeHandler)))

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}

}
