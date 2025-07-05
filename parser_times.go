package main

import (
	"fmt"
	"strings"

	"github.com/gocolly/colly/v2"
)

func translateTimesPage(url string, collector *colly.Collector) ([][]string, string, error) {
	ctx := &timesCtx{}
	registerTimesHandlers(collector, ctx)

	if err := collector.Visit(url); err != nil {
		return nil, "", fmt.Errorf("failed to visit URL, probably expired:%w", err)
	}

	driverTimes := translateRawLaptimes(ctx.rawDriverTimes)

	writables := translateTimesToCSV(driverTimes)

	return writables, fileNameFromTitleData(ctx.titleData, "LapTimes"), nil
}

func translateTimesToCSV(times []*DriverTimes) [][]string {
	if len(times) == 0 {
		return nil
	}

	// times * laps per driver
	final := make([][]string, 0, len(times)*16)

	final = append(final, driverTimeFields())

	for _, driverTime := range times {
		var (
			lapNum = 1
			rem    = driverTime.LapTimes
			meta   = lapMeta{}
		)

		for {
			meta, rem = getLapTimeSet(rem)
			final = append(final, driverTime.lapValues(meta, lapNum))

			if len(rem) == 0 {
				break
			}

			lapNum++
		}
	}

	return final
}

type lapMeta struct {
	s1, s2, s3, t string
}

func getLapTimeSet(rem []string) (lapMeta, []string) {
	var (
		s1, s2, s3, t string
		cutoff        = 0
	)

outer:
	for i, entry := range rem {
		cutoff = i

		if isPitstopData(entry) {
			entry = "PIT"
		} else if isNotCountedTime(entry) {
			entry = "NC"
		}

		switch i {
		case 0:
			s1 = entry
		case 1:
			s2 = entry
		case 2:
			s3 = entry
		case 3:
			t = entry
		default:
			break outer
		}
	}

	if len(rem) == 4 {
		rem = nil
	} else {
		rem = rem[cutoff:]
	}

	return lapMeta{s1, s2, s3, t}, rem
}

func registerTimesHandlers(collector *colly.Collector, ctx *timesCtx) {
	collector.OnHTML("table", func(h *colly.HTMLElement) {
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
				ctx.rawDriverTimes = splitDriverLaps(h.Text)
			}
		})
	})
}

func splitDriverLaps(s string) []string {
	splittage := strings.Split(s, "\n\n")
	if len(splittage) == 0 {
		return nil
	}

	// trim off the first entry which is just a line of underscores
	splittage = splittage[1:]

	end := 0
	for i, split := range splittage {
		if strings.Contains(strings.ToLower(split), "fastest sector") {
			end = i
			break
		}
	}

	return splittage[:end]
}

func translateRawLaptimes(rawTimes []string) []*DriverTimes {
	out := make([]*DriverTimes, 0, len(rawTimes))

	for _, rawData := range rawTimes {
		if strings.HasPrefix(rawData, "\n ") || strings.HasPrefix(rawData, "\n") {
			rawData = strings.TrimLeft(rawData, "\n ")
		}

		driverNumName, timingData, ok := strings.Cut(rawData, "\n")
		if !ok {
			continue
		}

		driverTimes := &DriverTimes{}

		num, name, _ := strings.Cut(strings.TrimSpace(driverNumName), " ")
		driverTimes.DriverNumber = clean(num)
		driverTimes.DriverName = clean(name)

		// this will get broken by the below if we don't replace here
		timingData = strings.ReplaceAll(timingData, "*:**.****", "NOT-COUNTED")

		// we dont care about picking out fastest lap data (indicated by a *)
		timingData = strings.ReplaceAll(timingData, "*", " ")

		for split := range strings.SplitSeq(timingData, " ") {
			if !isTimingData(split) {
				continue
			}

			split = strings.TrimSpace(split)

			driverTimes.LapTimes = append(driverTimes.LapTimes, split)
		}

		out = append(out, driverTimes)
	}

	return out
}

func isTimingData(s string) bool {
	//  0:50.8840 || -:--.----p || *:**.****
	return strings.Contains(s, ":") && strings.Contains(s, ".") || isNotCountedTime(s) || isPitstopData(s)
}

func isPitstopData(s string) bool {
	return strings.Contains(s, "p") && strings.Contains(s, "-")
}

func isNotCountedTime(s string) bool {
	if s == "NOT-COUNTED" {
		fmt.Println("what da helly")
		return true
	}
	return false
}
