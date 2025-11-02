package api

import "strconv"

type DriverTimes struct {
	DriverNumber string
	DriverName   string
	// ordered by lap
	LapTimes []string
}

type timesCtx struct {
	titleData      string
	rawDriverTimes []string
}

func (dt *DriverTimes) lapValues(meta lapMeta, lapNum int) []string {
	return []string{
		dt.DriverNumber, dt.DriverName,
		strconv.Itoa(lapNum), meta.t,
		meta.s1, meta.s2, meta.s3,
	}
}

func driverTimeFields() []string {
	return []string{
		"Driver Number", "Driver Name",
		"Lap Number", "Total Lap Time",
		"Sector 1 Time", "Sector 2 Time", "Sector 3 Time",
	}
}
