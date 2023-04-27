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

func handler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		w.Header().Add("Allow", "POST")
		http.Error(w, fmt.Sprintf("The method %s is not allowed. The POST method is allowed.", r.Method), http.StatusMethodNotAllowed)
		return
	}

	isGauge, err := regexp.MatchString(`/update/gauge/`, r.URL.RequestURI())
	if isGauge {
		gaugeHandler(w, r)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	isCounter, err := regexp.MatchString(`/update/counter/`, r.URL.RequestURI())
	if isCounter {
		counterHandler(w, r)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Error(w, "Not implemented", http.StatusNotImplemented)

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
	mux.Handle(`/update/`, http.HandlerFunc(handler))

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}

}
