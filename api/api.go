package api

import (
	"encoding/csv"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/gocolly/colly/v2"
)

var log *slog.Logger

func init() {
	log = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

type Parser interface {
	Parse(url string) (csvData [][]string, fileName string, err error)
	WriteEntriesTo(csvData [][]string, fileName string) error
}

func NewParser() Parser {
	return &parserClient{
		collector: colly.NewCollector(colly.Async(false)),
	}
}

type parserClient struct {
	collector *colly.Collector
}

func (p *parserClient) Parse(url string) (csvData [][]string, fileName string, err error) {
	// Detect format type
	switch true {
	case strings.Contains(url, "Result"):
		csvData, fileName, err = p.translateResultPage(url)
	case strings.Contains(url, "Times"):
		csvData, fileName, err = p.translateTimesPage(url)
	case strings.Contains(url, "127.0.0.1"):
		csvData, fileName, err = p.translateTimesPage(url)
	default:
		return nil, "", fmt.Errorf("unhandled page type based on url:%s", url)
	}

	if err != nil {
		log.Error("failed to translate, skipping", "URL", url, "error", err)
		return nil, "", err
	}

	return csvData, fileName, nil
}

func (p *parserClient) WriteEntriesTo(csvData [][]string, fileName string) error {
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := file.Truncate(0); err != nil {
		return err
	}

	writer := csv.NewWriter(file)

	return writer.WriteAll(csvData)
}

func isTitleData(s string) bool {
	return strings.Contains(strings.ToLower(s), "race") ||
		strings.Contains(strings.ToLower(s), "qual") ||
		strings.Contains(strings.ToLower(s), "prac")
}

func fileNameFromTitleData(s, dataType string) string {
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, " ", "")
	return s + "_" + dataType + ".csv"
}
