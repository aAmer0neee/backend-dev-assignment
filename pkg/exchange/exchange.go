package exchange

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aAmer0neee/backend-dev-assignment/types"
	"golang.org/x/net/html/charset"
)

const (
	url    = "http://www.cbr.ru/scripts/XML_daily_eng.asp?date_req="
	layout = "02/01/2006"
)

type DailyRate struct {
	XMLName xml.Name         `xml:"ValCurs"`
	Date    types.CustomDate `xml:"Date,attr"`
	Valutes []*Valute        `xml:"Valute"`
}

type Valute struct {
	Id        string            `xml:"ID,attr"`
	CharCode  string            `xml:"CharCode"`
	Nominal   int               `xml:"Nominal"`
	Name      string            `xml:"Name"`
	Value     types.CustomFloat `xml:"Value"`
	VunitRate types.CustomFloat `xml:"VunitRate"`
}

type ExchangeInfo struct {
	ListValutes []*DailyRate
	period      int

	MaxRate *Valute
	MaxDate time.Time
	MinRate *Valute
	MinDate time.Time

	AverageRub map[string]float64

	ch chan *DailyRate
}

func New(p int) *ExchangeInfo {
	return &ExchangeInfo{
		ListValutes: make([]*DailyRate, 0),
		period:      p,
		MinRate:     nil,
		MaxRate:     nil,
		AverageRub:  make(map[string]float64),
		ch:          make(chan *DailyRate),
	}
}

func (e *ExchangeInfo) Exchange() {

	go func() {

		if err := e.collectRates(); err != nil {
			log.Fatalf("ошибка парсинга страницы %v", err)
		}
	}()

	e.processRates()

	e.printExchangeInfo()
}

func (e *ExchangeInfo) collectRates() error {
	today := time.Now()
	endDate := time.Now().AddDate(0, 0, -e.period)

	dates := make(map[time.Time]bool)
	for today.After(endDate) {
		rate, err := e.parsePage(url + today.Format(layout))
		if err != nil {
			return err
		}
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
	return nil

}

func (e *ExchangeInfo) processRates() {
	for rate := range e.ch {
		e.checkStats(rate)

		e.ListValutes = append(e.ListValutes, rate)
	}

}

func (e *ExchangeInfo) parsePage(pageUrl string) (*DailyRate, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", pageUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	decoder := xml.NewDecoder(resp.Body)
	decoder.CharsetReader = charset.NewReaderLabel

	var valCurs DailyRate
	err = decoder.Decode(&valCurs)
	if err != nil {
		return nil, err
	}

	return &valCurs, nil
}

func (e *ExchangeInfo) checkStats(rate *DailyRate) {
	for _, valute := range rate.Valutes {

		value := float64(valute.VunitRate)

		e.AverageRub[valute.Id] += value

		if e.MinRate == nil || newIsLess(e.MinRate, value) {
			e.MinRate = valute
			e.MinDate = time.Time(rate.Date)
		}
		if e.MaxRate == nil || !newIsLess(e.MaxRate, value) {
			e.MaxRate = valute
			e.MaxDate = time.Time(rate.Date)
		}

	}
}

func newIsLess(old *Valute, new float64) bool {
	return new < float64(old.VunitRate)
}

func (e *ExchangeInfo) printExchangeInfo() {

	e.printAverage()

	fmt.Printf("%-40s%40s\n", "Максимальный курс:", "Минимальный курс:")

	left := fmt.Sprintf("Дата: %s", e.MaxDate.Format(layout))
	right := fmt.Sprintf("Дата: %s", e.MinDate.Format(layout))
	fmt.Printf("%-40s%40s\n", left, right)

	left = fmt.Sprintf("Валюта: %s (%s)", e.MaxRate.Name, e.MaxRate.CharCode)
	right = fmt.Sprintf("Валюта: %s (%s)", e.MinRate.Name, e.MinRate.CharCode)
	fmt.Printf("%-40s%40s\n", left, right)

	left = fmt.Sprintf("Курс за 1 %s: %.6f RUB", e.MaxRate.CharCode, float64(e.MaxRate.VunitRate))
	right = fmt.Sprintf("Курс за 1 %s: %.6f RUB", e.MinRate.CharCode, float64(e.MinRate.VunitRate))
	fmt.Printf("%-40s%40s\n\n", left, right)
}

func (e *ExchangeInfo) printAverage() {
	proc := make(map[string]bool)

	for _, list := range e.ListValutes {
		for _, v := range list.Valutes {
			if proc[v.Id] {
				continue
			}

			fmt.Printf("%-20s %-5s: %.6f RUB\n", v.Name, v.CharCode, e.AverageRub[v.Id])

			proc[v.Id] = true
		}
	}

	fmt.Println()
}
