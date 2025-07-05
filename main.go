package main

import (
	"bufio"
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

func main() {
	fmt.Printf("Paste one URL at a time and press enter, type s and then enter to stop adding URLs run the script\n-------------\n")
	urls := []string{}

	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		text := s.Text()

		if text == "" {
			continue
		}

		if text == "s" {
			break
		}

		urls = append(urls, text)
	}

	c := colly.NewCollector(colly.Async(false))

	// Detect format type
	for _, url := range urls {
		var (
			csvData  [][]string
			fileName string
			err      error
		)

		switch true {
		case strings.Contains(url, "Result"):
			csvData, fileName, err = translateResultPage(url, c)
		case strings.Contains(url, "Times"):
			csvData, fileName, err = translateTimesPage(url, c)
		default:
			log.Info("unhandled URL type", "URL", url)
			continue
		}

		if err != nil {
			log.Error("failed to translate, skipping", "URL", url, "error", err)
			continue
		}

		if err := writeCsv(csvData, fileName); err != nil {
			log.Info("Error writing CSV", "err", err.Error(), "filename", fileName)
			continue
		}

		log.Info("Successfully parsed results and saved file", "file name", fileName, "url", url)
	}
}

func writeCsv(entries [][]string, fileName string) error {
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := file.Truncate(0); err != nil {
		return err
	}

	return csv.NewWriter(file).WriteAll(entries)
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
