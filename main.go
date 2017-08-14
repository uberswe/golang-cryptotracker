package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"log"
	"io/ioutil"
	"strconv"
	"os/exec"
	"os"
	"runtime"
)

type CurrencyResult struct {
	Response string `json:"Response"`
	Message string `json:"Message"`
	BaseImageUrl string `json:"BaseImageUrl"`
	BaseLinkUrl string `json:"BaseLinkUrl"`
	Data Currency `json:"Data"`
	Type int `json:"Type"`
}

type Currency map[string]CurrencyDetails

type CurrencyDetails struct {
	Id string `json:"Id"`
	Url string `json:"Url"`
	ImageUrl string `json:"ImageUrl"`
	Name string `json:"Name"`
	CoinName string `json:"CoinName"`
	FullName string `json:"FullName"`
	Algorithm string `json:"Algorithm"`
	ProofType string `json:"ProofType"`
	FullyPremined string `json:"FullyPremined"`
	TotalCoinsFreeFloat string `json:"TotalCoinsFreeFloat"`
	TotalCoinSupply string `json:"TotalCoinSupply"`
	PreMinedValue string `json:"PreMinedValue"`
	SortOrder string `json:"SortOrder"`
}

type Config struct {
	Currencies []ConfigCurrency `json:"currencies"`
	Settings ConfigSettings `json:"settings"`
}

type ConfigCurrency map[string][]ConfigCurrencyDetails

type ConfigCurrencyDetails struct {
	Amount float64 `json:"Amount"`
	AtValue float64 `json:"AtValue"`
	WithCurrency string `json:"WithCurrency"`
}

type ConfigSettings struct {
	UpdateInterval int `json:"UpdateInterval"`
	UpdateTimeout int `json:"UpdateTimeout"`
	ApiUrl string `json:"ApiUrl"`
	CoinListUrl string `json:"CoinListUrl"`
}

type Result map[string]float64

var checks map[string]ConfigCurrencyDetails
var updateInterval int
var donateString = "Donate some ETH if you find this useful: 0xB9Df510bE5Aaad76E558cc7BF41E6363f3944dfc"

var clear map[string]func() //create a map for storing clear funcs

func init() {
	clear = make(map[string]func()) //Initialize it
	clear["linux"] = func() {
		cmd := exec.Command("clear") //Linux example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	clear["darwin"] = func() {
		cmd := exec.Command("clear") //OSX example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	clear["windows"] = func() {
		cmd := exec.Command("cls") //Windows example it is untested, but I think its working
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func CallClear() {
	value, ok := clear[runtime.GOOS] //runtime.GOOS -> linux, windows, darwin etc.
	if ok { //if we defined a clear func for that platform:
		value()  //we execute it
	} else { //unsupported platform
		panic("Your platform is unsupported! I can't clear terminal screen :(")
	}
}

func main() {

	updateInterval = 10
	checks = make(map[string]ConfigCurrencyDetails)
	currencies := make(map[string]bool)

	fmt.Println("Cryptotracker started")
	fmt.Println(donateString)

	fmt.Println("Reading config.json")

	config, err := ioutil.ReadFile("./config.json")

	c := Config{}
	jsonErr := json.Unmarshal(config, &c)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	for _, el := range c.Currencies {
		for coin, detailArray := range el {
			for _, detail := range detailArray {
				fmt.Println("Own " + FloatToString(detail.Amount) + " " + coin + " at " + FloatToString(detail.AtValue) + " " + detail.WithCurrency)
				currencies[coin] = false
				checks[coin] = detail
			}
		}
	}

	fmt.Println("API Url: " + c.Settings.ApiUrl)
	fmt.Println("Coin list Url: " + c.Settings.CoinListUrl)
	fmt.Println("Update interval: " + strconv.Itoa(c.Settings.UpdateInterval))
	fmt.Println("Timeout: " + strconv.Itoa(c.Settings.UpdateTimeout))

	fmt.Println("Verifying coin list")

	url := c.Settings.CoinListUrl

	spaceClient := http.Client{
		Timeout: time.Second * time.Duration(c.Settings.UpdateTimeout), // Maximum of 2 secs
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("User-Agent", "golang-cryptotracker")

	res, getErr := spaceClient.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	result := CurrencyResult{}
	jsonErr2 := json.Unmarshal(body, &result)
	if jsonErr2 != nil {
		log.Fatal(jsonErr2)
	}

	for index, _ := range result.Data {
		for i, _ := range currencies {
			if i == index {
				currencies[i] = true
			}
		}
	}

	for i, am := range currencies {
		if !am {
			fmt.Println("WARNING: " + i + " could not be found on coin list")
		}
	}

	fmt.Println("Done")
	fmt.Println("Preparing to make first check... please wait")

	doEvery(time.Second*time.Duration(updateInterval), getPrice)
	updateInterval = c.Settings.UpdateInterval
}

func getPrice(t time.Time) {
	CallClear();
	fmt.Printf("\033[0;0H")
	fmt.Printf("\r" + donateString + "\n")
	endCur := ""
	totalval := 0.00
	totalbought := 0.00
	for coin, detail := range checks {
		From := coin
		To := detail.WithCurrency
		url := "https://min-api.cryptocompare.com/data/price?fsym=" + From + "&tsyms=" + To;
		w := http.Client{
			Timeout: time.Second * 30,
		}

		req, _ := http.NewRequest(http.MethodGet, url, nil)

		res, getErr := w.Do(req)
		if getErr != nil {
			log.Println(getErr)
		} else {

			body, readErr := ioutil.ReadAll(res.Body)
			if readErr != nil {
				log.Println(readErr)
			} else {

				r := Result{}
				jsonErr := json.Unmarshal(body, &r)
				if jsonErr != nil {
					log.Println(jsonErr)
				} else {

					for _, p := range r {
						fmt.Println(t.Format(time.RFC3339) + ": " + FloatToString(detail.Amount) + " "+From+" = "+FloatToString(p*detail.Amount)+" "+To+" " + FloatToString(Change(detail.AtValue, p)) + "%")
						endCur = To
						totalval = totalval + p*detail.Amount
						totalbought = totalbought + detail.Amount*detail.AtValue
					}
				}
			}
		}
	}
	fmt.Println("Total purchased: " + FloatToString(totalbought) + " " + endCur)
	fmt.Println("Total value: " + FloatToString(totalval) + " " + endCur + " Change: " + FloatToString(Change(totalbought, totalval)) + "%")
}

func doEvery(d time.Duration, f func(time.Time)) {
	for x := range time.Tick(d) {
		f(x)
	}
}

func FloatToString(input_num float64) string {
	// to convert a float number to a string
	return strconv.FormatFloat(input_num, 'f', 6, 64)
}

func Change(before float64, after float64) float64{
	diff := after - before
	realDiff := diff / float64(before)
	percentDiff := 100 * realDiff

	return percentDiff
}