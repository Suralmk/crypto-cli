package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"unicode"

	"github.com/manifoldco/promptui"
)

type BinanceData struct {
	Price string `json:"price"`
}
type BitGetData struct {
	Data []struct {
		Price string `json:"lastpr"`
	}
}
type ErrorResponse struct {
	Msg string `json:"msg"`
}

func main() {
	// fetches the price from the selected exchange, and displays it.
	// The loop runs indefinitely until the program is terminated manually.
	for {
		validate := func(input string) error {
			if input == "" {
				return errors.New("input is empty")
			}
			for _, char := range input {
				if !unicode.IsUpper(char) {
					return errors.New("only capital letters are allowed")
				}
			}
			return nil
		}

		prompt := promptui.Prompt{
			Label:    "COIN SYMBOL",
			Validate: validate,
		}

		result, err := prompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}
		selectPrompt := promptui.Select{
			Label: "Select Exchange",
			Items: []string{"Binance", "Bitget"},
		}
		_, exchange, err := selectPrompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		switch exchange {
		case "Binance":
			price, err := FetchBinancePrice(result)
			if err != nil {
				fmt.Println("Error:", err)

			}
			DisplayPrice(result, price)
		case "Bitget":
			price, err := FetchBitGetPrice(result)
			if err != nil {
				fmt.Println("Error:", err)
			}
			DisplayPrice(result, price)
		}
	}
}

/*
FetchBinancePrice fetches the latest USDT price for a given cryptocurrency symbol from the Binance API.

Parameters:
  - symbol: The uppercase string representing the cryptocurrency symbol (e.g., "BTC").

Returns:
  - float64: The latest price of the given symbol in USDT.
  - error: An error object if the HTTP request fails, response is invalid, or parsing fails.
*/
func FetchBinancePrice(symbol string) (float64, error) {

	url := fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%sUSDT", symbol)
	res, err := http.Get(url)

	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return 0, checkAPIError(res)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	var price BinanceData
	err = json.Unmarshal(body, &price)
	if err != nil {
		panic(err)
	}
	f, _ := strconv.ParseFloat(price.Price, 64)
	return f, nil

}

/*
FetchBitGetPrice fetches the latest USDT price for a given cryptocurrency symbol from the bitget API.

Parameters:
  - symbol: The uppercase string representing the cryptocurrency symbol (e.g., "BTC").

Returns:
  - float64: The latest price of the given symbol in USDT.
  - error: An error object if the HTTP request fails, response is invalid, or parsing fails.
*/
func FetchBitGetPrice(symbol string) (float64, error) {
	url := fmt.Sprintf("https://api.bitget.com/api/v2/spot/market/tickers?symbol=%sUSDT", symbol)
	res, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return 0, checkAPIError(res)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	var price BitGetData
	err = json.Unmarshal(body, &price)
	if err != nil {
		panic("Failed to parse bitget data")
	}

	f, _ := strconv.ParseFloat(price.Data[0].Price, 64)
	return f, nil
}

func DisplayPrice(symbol string, price float64) {
	if price <= 0 {
		fmt.Printf("Price for %s is unavailable or invalid.\n", symbol)
		return
	}
	fmt.Printf("📈  %s-USDT -> $%.6f\n", symbol, price)
}

/*
checkAPIError reads the HTTP response body to extract API error messages.
Parameters:
  - res: Pointer to http.Response containing the error body.

Returns:
  - error: An error with the extracted API message, or a generic error if parsing fails.
*/
func checkAPIError(res *http.Response) error {
	errBody, _ := io.ReadAll(res.Body)
	var errResp ErrorResponse
	if err := json.Unmarshal(errBody, &errResp); err != nil {
		return errors.New("failed to parse error response")
	}
	return errors.New(errResp.Msg)
}
