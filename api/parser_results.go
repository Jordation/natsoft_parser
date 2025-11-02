package api

import (
	"fmt"
	"slices"
	"strings"

	"github.com/gocolly/colly/v2"
)

func (p *parserClient) translateResultPage(url string) ([][]string, string, error) {
	ctx := &resultsCtx{}
	registerHandlers(p.collector, ctx)

	if err := p.collector.Visit(url); err != nil {
		return nil, "", fmt.Errorf("failed to visit URL, err reason: URL %w", err)
	}

	fields, fieldLine := parseRawFields(ctx.rawData)

	stats := getStatsLines(ctx.rawData, fieldLine)

	results := convertRawResults(fieldLine, fields, stats)

	return translateResultsToCSV(results), fileNameFromTitleData(ctx.titleData, "Results"), nil
}

func parseRawFields(rawData string) ([]*resultField, string) {
	res := []*resultField{}
	lines := strings.Split(rawData, "\n")
	if len(lines) == 0 {
		log.Info("PARSER_ERR: no newline to split raw timing data", "RawData", rawData)
	}

	// garbage trimming
	firstLine := ""
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), "pos") &&
			strings.Contains(strings.ToLower(line), "car") &&
			strings.Contains(strings.ToLower(line), "driver") {
			firstLine = line
			break
		}
	}

	fieldName := ""
	fieldStart := 0
	for i, char := range firstLine {
		// reset if we run into new whitespace or hit the end of the line
		if fieldName != "" && char == ' ' {
			res = append(res, &resultField{
				firstCharIndex: fieldStart,
				fieldName:      fieldName,
			})

			fieldStart = 0
			fieldName = ""
		}

		if fieldName == "" && char != ' ' {
			fieldStart = i
		}

		if char != ' ' {
			fieldName += string(char)
		}
	}

	res = append(res, &resultField{
		firstCharIndex: fieldStart,
		fieldName:      fieldName,
		lastField:      true,
	})

	return res, firstLine
}

func getStatsLines(raw, fieldsLine string) []string {
	rawLines := []string{}
	outLines := []string{}

	splits := strings.Split(raw, "\n")
	for i, split := range splits {
		if split == fieldsLine {
			rawLines = splits[i+2:]
			break
		}
	}

	for _, rawLine := range rawLines {
		if rawLine == "" {
			break
		}

		outLines = append(outLines, rawLine)
	}

	return outLines
}

func convertRawResults(header string, rawFields []*resultField, rawStats []string) []*Results {
	res := []*Results{}
	for _, rawStatLine := range rawStats {
		rslt := &Results{}

		lastChar := 0
		for j, f := range rawFields {
			if f.lastField {
				lastChar = len(rawStatLine)
			} else {
				lastChar = rawFields[j+1].firstCharIndex
			}

			// if the stat line is shorter we can assume the following stats are missing and just set it like so
			if lastChar-1 > len(rawStatLine) {
				lastChar = len(rawStatLine) - 1
			}

			if f.firstCharIndex > len(rawStatLine) {
				continue
			}

			statValue := strings.TrimSpace(rawStatLine[f.firstCharIndex : lastChar-1])
			assignResults(f.fieldName, statValue, rslt)
		}

		res = append(res, rslt)
	}

	return res
}

func assignResults(fieldName, value string, to *Results) {
	cleaned := clean(value)
	if cleaned == "" {
		return
	}

	switch fieldName {
	case "Pos":
		to.Position = cleaned
	case "Car":
		to.CarNumber = cleaned
	case "Driver":
		to.Driver = cleaned
	case "Competitor/Team":
		to.Team = cleaned
	case "Vehicle":
		to.Car = cleaned
	case "CL":
		to.Class = cleaned
	case "Laps":
		to.Laps = cleaned
	case "Race.Time":
		to.RaceTime = cleaned
	case "Fastest...Lap":
		to.FastestLap = cleaned
	case "Cap":
		to.Capacity = cleaned
	case "Gap":
		to.Gap = cleaned
	default:
		log.Info("unhandled field", "fieldName", fieldName, "cleaned value", cleaned)
	}
}

func clean(s string) string {
	out := []string{}

	for split := range strings.SplitSeq(s, ".") {
		clean := strings.TrimSpace(split)
		out = append(out, clean)
	}

	return strings.Join(out, " ")
}

func registerHandlers(c *colly.Collector, ctx *resultsCtx) {
	c.OnHTML("table", func(h *colly.HTMLElement) {
		h.ForEach("tr", func(i int, h *colly.HTMLElement) {
			switch i {
			// case title:
			case 0:
				splits := strings.Split(h.Text, "\n")
				for _, split := range splits {
					if isTitleData(split) {
						ctx.titleData = split
						break
					}
				}
			// case data:
			case 2:
				ctx.rawData = h.Text
			}
		})
	})
}

func translateResultsToCSV(results []*Results) [][]string {
	if len(results) == 0 {
		return nil
	}

	// [cap, cap, cap]
	// [carN, carN, carN].. etc
	validDatas := make([][]string, len(resultsFields()))
	for _, result := range results {
		for i, stat := range result.values() {
			if stat != "" {
				validDatas[i] = append(validDatas[i], stat)
			}
		}
	}

	discard := []int{}
	for fieldIndex, stats := range validDatas {
		// if we had less than a quater of entries with this stat, we don't want it!
		if len(stats) < len(results)/4 {
			discard = append(discard, fieldIndex)
		}
	}

	headers := make([]string, 0, len(results)-len(discard))
	for fieldIndex, field := range resultsFields() {
		if !slices.Contains(discard, fieldIndex) {
			headers = append(headers, field)
		}
	}

	final := make([][]string, 0, len(results)+1) // +header
	final = append(final, headers)

	for _, result := range results {
		inter := make([]string, 0, len(headers))
		for fieldIndex, val := range result.values() {
			if !slices.Contains(discard, fieldIndex) {
				inter = append(inter, val)
			}
		}

		final = append(final, inter)
	}

	return final
}
