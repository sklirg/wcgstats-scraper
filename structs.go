package main

import "encoding/xml"

type StatisticsHistory struct {
	XMLName               xml.Name `xml:"StatisticsHistory" json:"StatisticsHistory"`
	DailyStatisticsTotals []DailyStatisticsTotals
}

type DailyStatisticsTotals struct {
	Date    string `xml:"Date"`
	RunTime int64  `xml:"RunTime"`
	Points  int64  `xml:"Points"`
	Results int64  `xml:"Results"`
}
