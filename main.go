package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
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
}

var funds = []vgfund{
	{Name: "LifeStrategy 20% Equity", URL: "https://www.vanguardinvestor.co.uk/api/funds/vanguard-lifestrategy-20-equity-fund-gbp-gross-accumulation-shares"},
	{Name: "LifeStrategy 40% Equity", URL: "https://www.vanguardinvestor.co.uk/api/funds/vanguard-lifestrategy-40-equity-fund-accumulation-shares"},
	{Name: "LifeStrategy 60% Equity", URL: "https://www.vanguardinvestor.co.uk/api/funds/vanguard-lifestrategy-60-equity-fund-accumulation-shares"},
	{Name: "LifeStrategy 80% Equity", URL: "https://www.vanguardinvestor.co.uk/api/funds/vanguard-lifestrategy-80-equity-fund-accumulation-shares"},
	{Name: "LifeStrategy 100% Equity", URL: "https://www.vanguardinvestor.co.uk/api/funds/vanguard-lifestrategy-100-equity-fund-accumulation-shares"},
}

func main() {
	wg := &sync.WaitGroup{}
	wg.Add(len(funds))

	for _, fund := range funds {
		go retrieve(fund, wg)
	}

	wg.Wait()
}

func retrieve(fund vgfund, wg *sync.WaitGroup) {
	defer wg.Done()

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
	fmt.Printf("%s: %s\n", fund.Name, d.NavPrice.Value)
}
