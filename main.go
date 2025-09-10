package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type vgfund struct {
	Name  string
	URL   string
	Price float64
}

type result struct {
	vgfund
	err     error
	latency time.Duration
}

var availableFunds = map[string]vgfund{
	"ls20":  {Name: "LifeStrategy 20% Equity", URL: "https://www.vanguardinvestor.co.uk/api/funds/vanguard-lifestrategy-20-equity-fund-gbp-gross-accumulation-shares"},
	"ls40":  {Name: "LifeStrategy 40% Equity", URL: "https://www.vanguardinvestor.co.uk/api/funds/vanguard-lifestrategy-40-equity-fund-accumulation-shares"},
	"ls60":  {Name: "LifeStrategy 60% Equity", URL: "https://www.vanguardinvestor.co.uk/api/funds/vanguard-lifestrategy-60-equity-fund-accumulation-shares"},
	"ls80":  {Name: "LifeStrategy 80% Equity", URL: "https://www.vanguardinvestor.co.uk/api/funds/vanguard-lifestrategy-80-equity-fund-accumulation-shares"},
	"ls100": {Name: "LifeStrategy 100% Equity", URL: "https://www.vanguardinvestor.co.uk/api/funds/vanguard-lifestrategy-100-equity-fund-accumulation-shares"},
}

func main() {
	err := run(os.Args, os.Stdin, os.Stdout, os.Stderr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	selectedFunds := getSelectedFunds()

	client := &http.Client{Timeout: 10 * time.Second}

	ch := make(chan result, len(selectedFunds))
	for _, fund := range selectedFunds {
		go retrieveFundPrice(client, fund, ch)
	}

	for i := len(selectedFunds); i != 0; i-- {
		res := <-ch
		if res.err != nil {
			fmt.Fprintln(stdout, "Failed to receive fund info for", res.Name, res.err)
			continue
		}

		fmt.Fprintf(stdout, "%-25s £%.2f (%s)\n", res.Name, res.Price, res.latency)
	}

	return nil
}

func getSelectedFunds() []vgfund {
	var selectedFunds []vgfund

	if len(os.Args) == 2 {
		cliFunds := strings.Split(os.Args[1], ",")

		for _, v := range cliFunds {
			fund, ok := availableFunds[v]
			// check map key existence before adding to selected funds
			if ok {
				selectedFunds = append(selectedFunds, fund)
			} else {
				fmt.Printf("Fund with ID %s not recognised\n", v)
			}
		}
	} else {
		for _, v := range availableFunds {
			selectedFunds = append(selectedFunds, v)
		}
	}
	return selectedFunds
}

func retrieveFundPrice(client *http.Client, fund vgfund, ch chan<- result) {
	start := time.Now()

	resp, err := client.Get(fund.URL)
	if err != nil {
		ch <- result{fund, errors.New("failed to get fund info from URL"), time.Since(start).Round(time.Millisecond)}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		ch <- result{fund, errors.New("bad response from vanguard"), time.Since(start).Round(time.Millisecond)}
	}

	var d struct {
		NavPrice struct {
			Value string `json:"value"`
		} `json:"navPrice"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		ch <- result{fund, errors.New("failed to decode response body"), time.Since(start).Round(time.Millisecond)}
		return
	}

	fundPrice, err := strconv.ParseFloat(d.NavPrice.Value, 64)
	if err != nil {
		ch <- result{fund, errors.New("failed to read NavPrice value as float"), time.Since(start).Round(time.Millisecond)}
		return
	}

	fund.Price = fundPrice

	ch <- result{fund, nil, time.Since(start).Round(time.Millisecond)}
}
