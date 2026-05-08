package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rivo/tview"
)

const (
	MetricsURL      = "http://localhost:8080/metrics"
	RefreshInterval = 1 * time.Second
	GridRowSize     = 3
	GridColSize     = 40
)

type Metrics struct {
	Requests int `json:"requests_total"`
	Errors   int `json:"errors_total"`
}

func fetchMetrics(client *http.Client) Metrics {
	resp, err := client.Get(MetricsURL)
	if err != nil {
		return Metrics{}
	}
	defer resp.Body.Close()

	var m Metrics
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return Metrics{}
	}
	return m
}

func createTextView(initialText string) *tview.TextView {
	return tview.NewTextView().
		SetDynamicColors(true). // Allows using [red], [green] tags
		SetTextAlign(tview.AlignCenter).
		SetText(initialText)
}

func main() {

	app := tview.NewApplication()

	requests := createTextView("Requests: ...")
	errors := createTextView("Errors: ...")

	grid := tview.NewGrid().
		SetRows(GridRowSize, GridRowSize).
		SetColumns(GridColSize, GridColSize).
		SetBorders(true).
		AddItem(requests, 0, 0, 1, 1, 0, 0, false).
		AddItem(errors, 0, 1, 1, 1, 0, 0, false)

	go func() {
		ticker := time.NewTicker(RefreshInterval)
		defer ticker.Stop()

		client := &http.Client{Timeout: 2 * time.Second}

		for range ticker.C {
			m := fetchMetrics(client)

			// Safely update the UI from a goroutine
			app.QueueUpdateDraw(func() {
				requests.SetText(fmt.Sprintf("Requests: [green]%d", m.Requests))
				errors.SetText(fmt.Sprintf("Errors: [red]%d", m.Errors))
			})
		}
	}()

	if err := app.SetRoot(grid, true).Run(); err != nil {
		panic(err)
	}
}
