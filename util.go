package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func ThisWeek() string {
	date := time.Now()
	for date.Weekday() != time.Monday {
		date = date.AddDate(0, 0, -1)
	}
	end := date.AddDate(0, 0, 6)
	return fmt.Sprintf("%s - %s", date.Format("01/02"), end.Format("01/02"))
}

func HTTPDownload(uri string) ([]byte, error) {
	fmt.Printf("HTTPDownload From: %s.\n", uri)
	res, err := http.Get(uri)
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()
	d, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("ReadFile: Size of download: %d\n", len(d))
	return d, err
}
