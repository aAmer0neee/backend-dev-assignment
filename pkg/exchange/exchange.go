package exchange

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html/charset"
)

const (
	url    = "http://www.cbr.ru/scripts/XML_daily_eng.asp?date_req="
	layout = "02/01/2006"
)

type DailyRate struct {
	XMLName xml.Name   `xml:"ValCurs"`
	Date    CustomDate `xml:"Date,attr"`
	Valutes []*Valute  `xml:"Valute"`
}

type Valute struct {
	Id        string      `xml:"ID,attr"`
	CharCode  string      `xml:"CharCode"`
	Nominal   int         `xml:"Nominal"`
	Name      string      `xml:"Name"`
	Value     CustomFloat `xml:"Value"`
	VunitRate CustomFloat `xml:"VunitRate"`
}

type CustomDate time.Time
type CustomFloat float64

func (f *CustomFloat) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var content string
	if err := d.DecodeElement(&content, &start); err != nil {
		return err
	}
	value, err := strconv.ParseFloat(strings.Replace(content, ",", ".", 1), 64)
	if err != nil {
		return err
	}
	*f = CustomFloat(value)
	return nil
}
func (f *CustomFloat) Float64() float64 {
	return float64(*f)
}

func (d *CustomDate) UnmarshalXMLAttr(attr xml.Attr) error {

	date, err := time.Parse("02.01.2006", attr.Value)
	if err != nil {
		return err
	}
	*d = CustomDate(date)
	return nil
}
func (d CustomDate) String() string {
	return time.Time(d).Format("02.01.2006")
}

type ExchangeInfo struct {
	ListValutes []*DailyRate
	period int

	MaxRate *Valute
	MinRate *Valute

	AverageRub map[string]float64

	ch chan *DailyRate
}


func New(p int) *ExchangeInfo {
	return &ExchangeInfo{
		ListValutes: make([]*DailyRate, 0),
		period: p,
		MinRate:     nil,
		MaxRate:     nil,
		AverageRub:  make(map[string]float64),
		ch: make(chan *DailyRate),
	}
}

func (e *ExchangeInfo) Exchange() {

	go e.collectRates()

	e.processRates()
	
	printExchangeInfo(e)
}

func (e *ExchangeInfo)collectRates(){
	today := time.Now()
	endDate := time.Now().AddDate(0, 0, -e.period)

	dates := make(map[time.Time]bool)
	for today.After(endDate) {
		rate := e.parsePage(url + today.Format(layout))
		if dates[time.Time(rate.Date)] {
			today = today.AddDate(0, 0, -1)
			continue
		} else {
			dates[time.Time(rate.Date)] = true
		}

		e.ch <- rate

		today = today.AddDate(0, 0, -1)
	}
	close(e.ch)

}

func (e *ExchangeInfo)processRates(){
	for rate := range e.ch {
		e.checkStats(rate)
		
		e.ListValutes = append(e.ListValutes, rate)
	}
	
}

func (e *ExchangeInfo) parsePage(pageUrl string) *DailyRate {
	client := &http.Client{}
	req, err := http.NewRequest("GET", pageUrl, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	decoder := xml.NewDecoder(resp.Body)
	decoder.CharsetReader = charset.NewReaderLabel

	var valCurs DailyRate
	err = decoder.Decode(&valCurs)
	if err != nil {
		panic(err)
	}

	return &valCurs
}


func (e*ExchangeInfo)checkStats(rate *DailyRate){
	for _, valute := range rate.Valutes {

		value := float64(valute.VunitRate)

		e.AverageRub[valute.Id] += value

		if e.MinRate == nil || newIsLess(e.MinRate, value) {
			e.MinRate = valute
		}
		if e.MaxRate == nil || !newIsLess(e.MaxRate, value) {
			e.MaxRate = valute
		}

	}
}

func newIsLess(old *Valute, new float64) bool {
	return new < float64(old.VunitRate)
}

func printExchangeInfo(e *ExchangeInfo) {
	// Печать ListValutes
	for _, dailyRate := range e.ListValutes {
		fmt.Printf("Date: %s\n", dailyRate.Date)
		for _, valute := range dailyRate.Valutes {
			fmt.Printf("  Valute: %s, CharCode: %s, VunitRate: %f\n", valute.Name, valute.CharCode, valute.VunitRate)
		}
	}

	// Печать MaxRate
	fmt.Println("\nMax Rate:")
	printValute(e.MaxRate)

	// Печать MinRate
	fmt.Println("\nMin Rate:")
	printValute(e.MinRate)

	// Печать AverageRub
	fmt.Println("\nAverage Rub:")
	for id, avg := range e.AverageRub {
		fmt.Printf("Valute ID: %s, Average: %.4f\n", id, avg)
	}
}

// Функция для печати информации о валюте
func printValute(v *Valute) {
	if v != nil {
		fmt.Printf("ID: %s, CharCode: %s, Name: %s, VunitRate: %s\n", v.Id, v.CharCode, v.Name, v.VunitRate)
	} else {
		fmt.Println("Valute is nil.")
	}
}
