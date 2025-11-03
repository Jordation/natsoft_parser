package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"

	"github.com/Jordation/natsoftparser/api"
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

	parser := api.NewParser()

	for _, url := range urls {
		csvData, fileName, err := parser.Parse(url)
		if err != nil {
			log.Error("failed to translate, skipping", "URL", url, "error", err)
			continue
		}

		file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			log.Info("failed opening file:", "err", err.Error(), "filename", fileName)
			continue
		}
		defer file.Close()

		if err := file.Truncate(0); err != nil {
			log.Info("failed to truncate fail", "err", err.Error(), "filename", fileName)
			continue
		}

		if err := parser.WriteEntriesTo(csvData, file); err != nil {
			log.Info("Error writing CSV", "err", err.Error(), "filename", fileName)
			continue
		}

		log.Info("Successfully parsed results and saved file", "file name", fileName, "url", url)
	}
}
