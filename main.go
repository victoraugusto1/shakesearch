package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"index/suffixarray"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	searcher := Searcher{}
	err := searcher.Load("completeworks.txt")
	if err != nil {
		log.Fatal(err)
	}

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/search", handleSearch(searcher))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	fmt.Printf("Listening on port %s...", port)
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		log.Fatal(err)
	}
}

type Searcher struct {
	CompleteWorks              string
	CaseSensitiveSuffixArray   *suffixarray.Index
	CaseInsensitiveSuffixArray *suffixarray.Index
}

func handleSearch(searcher Searcher) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query, ok := r.URL.Query()["q"]
		if !ok || len(query[0]) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("missing search query in URL params"))
			return
		}

		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)

		maxResults, err := searchMaxResults(r)
		if err != nil {
			enc.Encode([1]string{"Invalid value for max number of results"})
			w.Write(buf.Bytes())
			return
		}

		caseSensitive := checkCaseSensitive(r)

		results := searcher.Search(query[0], caseSensitive, maxResults)
		if len(results) == 0 {
			enc.Encode([1]string{"Your search did not match any results"})
			w.Write(buf.Bytes())
			return
		}

		encodingErr := enc.Encode(results)
		if encodingErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("encoding failure"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(buf.Bytes())
	}
}

func (s *Searcher) Load(filename string) error {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Load: %w", err)
	}
	s.CompleteWorks = string(dat)
	s.CaseInsensitiveSuffixArray = suffixarray.New(bytes.ToLower(dat))
	s.CaseSensitiveSuffixArray = suffixarray.New(dat)
	return nil
}

func (s *Searcher) Search(query string, caseSensitive bool, maxResults int) []string {
	idxs := s.getSearchIndexes(query, caseSensitive, maxResults)
	results := []string{}
	for _, idx := range idxs {
		results = append(results, s.CompleteWorks[idx-250:idx+250])
	}
	return results
}

func (s *Searcher) getSearchIndexes(query string, caseSensitive bool, maxResults int) []int {
	if caseSensitive == true {
		return s.CaseSensitiveSuffixArray.Lookup([]byte(query), maxResults)
	}
	lowercaseByteQuery := bytes.ToLower([]byte(query))
	return s.CaseInsensitiveSuffixArray.Lookup(lowercaseByteQuery, maxResults)
}

func checkCaseSensitive(r *http.Request) bool {
	_, err := r.URL.Query()["caseSensitive"]
	caseSensitive := true
	if !err {
		caseSensitive = false
	}
	return caseSensitive
}

func searchMaxResults(r *http.Request) (int, error) {
	maxResultsQuery, ok := r.URL.Query()["maxResults"]
	if !ok || len(maxResultsQuery[0]) < 1 {
		return -1, nil
	}
	maxResults, err := strconv.Atoi(maxResultsQuery[0])
	if err != nil {
		return 0, fmt.Errorf("Invalid value for max number of results")
	}
	return maxResults, nil
}
