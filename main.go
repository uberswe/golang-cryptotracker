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
	"github.com/fatih/color"
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
}

type CurrencyObj struct {
	Valid bool
	Currency string
	Positions []ConfigCurrencyDetails
}

type ConfigSettings struct {
	UpdateInterval int `json:"UpdateInterval"`
	UpdateTimeout int `json:"UpdateTimeout"`
	ApiUrl string `json:"ApiUrl"`
	CoinListUrl string `json:"CoinListUrl"`
	BaseCurrency string `json:"BaseCurrency"`
	Color bool `json:"Color"`
	ShowConversion bool `json:"ShowConversion"`
}

type Result map[string]float64

var checks []CurrencyObj
var updateInterval int
var updateTimeout int
var baseCurrency string
var donateString = "Donate some ETH if you find this useful: 0xB9Df510bE5Aaad76E558cc7BF41E6363f3944dfc"
var colorOn bool
var showConversion bool

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

	fmt.Println("Cryptotracker started")
	fmt.Println(donateString)

	fmt.Println("Reading config.json")

	config, err := ioutil.ReadFile("./config.json")

	c := Config{}
	jsonErr := json.Unmarshal(config, &c)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	updateInterval = c.Settings.UpdateInterval
	baseCurrency = c.Settings.BaseCurrency
	updateTimeout = c.Settings.UpdateTimeout
	colorOn = c.Settings.Color
	showConversion = c.Settings.ShowConversion

	for _, el := range c.Currencies {
		for coin, detailArray := range el {
			curr := CurrencyObj{false, coin, []ConfigCurrencyDetails{}}
			for _, detail := range detailArray {
				fmt.Println("Own " + FloatToString(detail.Amount) + " " + coin + " at " + FloatToString(detail.AtValue) + " " + baseCurrency)
				curr.Positions = append(curr.Positions, detail)
			}
			checks = append(checks, curr)
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
		for _, obj := range checks {
			if obj.Currency == index {
				obj.Valid = true
			}
		}
	}

	for _, am := range checks {
		if !am.Valid {
			fmt.Println("WARNING: " + am.Currency + " could not be found on coin list")
		}
	}

	fmt.Println("Done")
	fmt.Println("Preparing to make first check... please wait")
	CallClear();

	getPrice(time.Now())
	doEvery(time.Second*time.Duration(updateInterval), getPrice)
}

func getPrice(t time.Time) {
	fmt.Printf("\033[0;0H")
	fmt.Printf("\r" + donateString + "\n")
	fmt.Println("Last updated: " + t.Format(time.RFC3339))
	endCur := ""
	totalval := 0.00
	totalbought := 0.00
	for _, detail := range checks {

		LastTo := ""
		LastFrom := ""
		LastP := 0.00

		for _, cur := range detail.Positions {

			From := detail.Currency
			To := baseCurrency
			url := "https://min-api.cryptocompare.com/data/price?fsym=" + From + "&tsyms=" + To;
			w := http.Client{
				Timeout: time.Second * time.Duration(updateTimeout),
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
							resString :=  FloatToString(cur.Amount) + " 	" + From + " 	= 	" + FloatToString(p*cur.Amount) + " 	" + To + "	 " + FloatToString(Change(cur.AtValue, p)) + "%"
							if colorOn {
								if Change(cur.AtValue, p) > 0.00 {
									c := color.New(color.FgGreen)
									c.Println(resString)
								} else {
									c := color.New(color.FgRed)
									c.Println(resString)
								}
							} else {
								fmt.Println(resString)
							}
							endCur = To
							totalval = totalval + p*cur.Amount
							totalbought = totalbought + cur.Amount*cur.AtValue
							LastTo = To
							LastFrom = From
							LastP = p
						}
					}
				}
			}
		}
		if showConversion {
			conversionString := "1 		" + LastFrom + " 	= 	" + FloatToString(LastP) + " 	" + LastTo
			if colorOn {
				c := color.New(color.FgBlue).Add(color.Bold)
				c.Println(conversionString)
			} else {
				fmt.Println(conversionString)
			}
		}
	}
	fmt.Println("Total purchased: " + FloatToString(totalbought) + " " + endCur)
	fmt.Println("Total value: " + FloatToString(totalval) + " " + endCur)
	fmt.Println("Change: " + FloatToString(Change(totalbought, totalval)) + "%")
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