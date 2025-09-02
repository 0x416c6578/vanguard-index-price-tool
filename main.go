package main

import (
	"encoding/json"
	"errors"
	"fmt"
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
	err error
}

var availableFunds = map[string]vgfund{
	"ls20":  {Name: "LifeStrategy 20% Equity", URL: "https://www.vanguardinvestor.co.uk/api/funds/vanguard-lifestrategy-20-equity-fund-gbp-gross-accumulation-shares"},
	"ls40":  {Name: "LifeStrategy 40% Equity", URL: "https://www.vanguardinvestor.co.uk/api/funds/vanguard-lifestrategy-40-equity-fund-accumulation-shares"},
	"ls60":  {Name: "LifeStrategy 60% Equity", URL: "https://www.vanguardinvestor.co.uk/api/funds/vanguard-lifestrategy-60-equity-fund-accumulation-shares"},
	"ls80":  {Name: "LifeStrategy 80% Equity", URL: "https://www.vanguardinvestor.co.uk/api/funds/vanguard-lifestrategy-80-equity-fund-accumulation-shares"},
	"ls100": {Name: "LifeStrategy 100% Equity", URL: "https://www.vanguardinvestor.co.uk/api/funds/vanguard-lifestrategy-100-equity-fund-accumulation-shares"},
}

func main() {
	selectedFunds := getSelectedFunds()

	client := &http.Client{Timeout: 10 * time.Second}

	ch := make(chan result, len(selectedFunds))
	for _, fund := range selectedFunds {
		go retrieveFundPrice(client, fund, ch)
	}

	for i := len(selectedFunds); i != 0; i-- {
		fundInfo := <-ch
		if fundInfo.err != nil {
			fmt.Println("Failed to receive fund info for", fundInfo.Name, fundInfo.err)
			continue
		}

		fmt.Println(fundInfo.Name, fundInfo.Price)
	}
}

func getSelectedFunds() []vgfund {
	var selectedFunds []vgfund

	if len(os.Args) == 2 {
		cliFunds := strings.Split(os.Args[1], ",")

		for _, v := range cliFunds {
			fund, ok := availableFunds[v]
			if ok {
				selectedFunds = append(selectedFunds, fund)
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
	resp, err := client.Get(fund.URL)
	if err != nil {
		ch <- result{fund, errors.New("failed to get fund info from URL")}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		ch <- result{fund, errors.New("bad response from vanguard")}
	}

	var d struct {
		NavPrice struct {
			Value string `json:"value"`
		} `json:"navPrice"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		ch <- result{fund, errors.New("failed to decode response body")}
		return
	}

	fundPrice, err := strconv.ParseFloat(d.NavPrice.Value, 64)
	if err != nil {
		ch <- result{fund, errors.New("failed to read NavPrice value as float")}
		return
	}

	fund.Price = fundPrice

	ch <- result{fund, nil}
}
