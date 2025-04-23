package types

import (
	"encoding/xml"
	"strconv"
	"strings"
	"time"
)

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
