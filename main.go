package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"

	"github.com/gocolly/colly/v2"
)

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
		)

		switch true {
		case strings.Contains(url, "Result"):
			csvData, fileName = translateResultPage(url, c)
		case strings.Contains(url, "Times"):
			csvData, fileName = translateTimesPage(url, c)
		default:
			log.Default().Printf("unhandled URL type, URL:%s", url)
			continue
		}

		if err := writeCsv(csvData, fileName); err != nil {
			confirmFatal("Error writing CSV:%s", err)
		}

		slog.Info("Successfully parsed %s results and saved to %s\n", url, fileName)
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

func debug(fmt string, args ...any) {
	slog.Debug("[DEBUG]: "+fmt, args...)
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

func confirmFatal(fmt string, args ...any) {
	slog.Info(fmt, args...)
	slog.Info("Fatal Error, press enter to exit...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
	os.Exit(1)
}
