package model

import "time"

type Settings struct {
	FileName       string `json:"fileName,omitempty"`
	Year           int    `json:"year,omitempty"`
	Width          int    `json:"width,omitempty"`
	Height         int    `json:"height,omitempty"`
	MarginLeft     int    `json:"marginLeft,omitempty"`
	MarginRight    int    `json:"marginRight,omitempty"`
	MarginTop      int    `json:"marginTop,omitempty"`
	MarginBottom   int    `json:"marginBottom,omitempty"`
	StartOfTheWeek int    `json:"startOfTheWeek,omitempty"`
	HeaderFontSize int    `json:"headerFontSize,omitempty"`
	HeaderFont     string `json:"headerFont,omitempty"`
}

func DefaultSettings() Settings {
	return Settings{
		Year:           time.Now().Year(),
		Width:          1404,
		Height:         1872,
		MarginLeft:     130,
		MarginRight:    10,
		MarginTop:      5,
		MarginBottom:   200,
		StartOfTheWeek: 1,
		HeaderFontSize: 25,
		HeaderFont:     "arial",
	}
}
