package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
)

// struct for incoming data, containing only used fields
type data struct {
	NavPrice struct {
		Value string `json:"value"`
	} `json:"navPrice"`
}

type vgfund struct {
	Name string
	URL  string
	ID   string
}

type vgfundprice struct {
	vgfund
	price float64
}

var availableFunds = []vgfund{
	{Name: "LifeStrategy 20% Equity", ID: "ls20", URL: "https://www.vanguardinvestor.co.uk/api/funds/vanguard-lifestrategy-20-equity-fund-gbp-gross-accumulation-shares"},
	{Name: "LifeStrategy 40% Equity", ID: "ls40", URL: "https://www.vanguardinvestor.co.uk/api/funds/vanguard-lifestrategy-40-equity-fund-accumulation-shares"},
	{Name: "LifeStrategy 60% Equity", ID: "ls60", URL: "https://www.vanguardinvestor.co.uk/api/funds/vanguard-lifestrategy-60-equity-fund-accumulation-shares"},
	{Name: "LifeStrategy 80% Equity", ID: "ls80", URL: "https://www.vanguardinvestor.co.uk/api/funds/vanguard-lifestrategy-80-equity-fund-accumulation-shares"},
	{Name: "LifeStrategy 100% Equity", ID: "ls100", URL: "https://www.vanguardinvestor.co.uk/api/funds/vanguard-lifestrategy-100-equity-fund-accumulation-shares"},
}

func main() {
	var selectedFunds []vgfund

	if len(os.Args) == 2 {
		cliFunds := strings.Split(os.Args[1], ",")

		for _, v := range availableFunds {
			if slices.Contains(cliFunds, v.ID) {
				selectedFunds = append(selectedFunds, v)
			}
		}
	} else {
		selectedFunds = availableFunds
	}

	ch := make(chan vgfundprice, len(selectedFunds))

	for _, fund := range selectedFunds {
		go retrieve(fund, ch)
	}

	for i := len(selectedFunds); i != 0; i-- {
		fundPrice := <-ch
		fmt.Println(fundPrice.Name, fundPrice.price)
	}

}

func retrieve(fund vgfund, ch chan vgfundprice) {
	resp, err := http.Get(fund.URL)
	if err != nil {
		fmt.Println("Couldn't retrieve %s", fund.URL)
		return
	}
	defer resp.Body.Close()

	var d data
	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return
	}

	fundValue, err := strconv.ParseFloat(d.NavPrice.Value, 64)

	ch <- vgfundprice{fund, fundValue}
}
