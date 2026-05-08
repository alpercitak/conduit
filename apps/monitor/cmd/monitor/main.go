package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rivo/tview"
)

type Metrics struct {
	Requests int `json:"requests_total"`
	Errors   int `json:"errors_total"`
}

func fetchMetrics() Metrics {
	resp, err := http.Get("http://localhost:8080/metrics")
	if err != nil {
		return Metrics{}
	}
	defer resp.Body.Close()

	var m Metrics
	json.NewDecoder(resp.Body).Decode(&m)
	return m
}

func main() {

	app := tview.NewApplication()

	requests := tview.NewTextView().
		SetText("Requests: ...")

	errors := tview.NewTextView().
		SetText("Errors: ...")

	grid := tview.NewGrid().
		SetRows(3, 3).
		SetColumns(40, 40).
		AddItem(requests, 0, 0, 1, 1, 0, 0, false).
		AddItem(errors, 0, 1, 1, 1, 0, 0, false)

	go func() {
		ticker := time.NewTicker(1 * time.Second)

		for range ticker.C {
			m := fetchMetrics()

			app.QueueUpdateDraw(func() {
				requests.SetText(fmt.Sprintf("Requests: %d", m.Requests))
				errors.SetText(fmt.Sprintf("Errors: %d", m.Errors))
			})
		}
	}()

	if err := app.SetRoot(grid, true).Run(); err != nil {
		panic(err)
	}
}
