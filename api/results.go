package api

type resultField struct {
	firstCharIndex int
	fieldName      string
	lastField      bool
}

type resultsCtx struct {
	titleData string
	rawData   string
}

type Results struct {
	Capacity   string
	CarNumber  string
	Car        string
	Class      string
	Driver     string
	FastestLap string
	Gap        string
	Laps       string
	LapTime    string
	Position   string
	RaceTime   string
	Team       string
}

func (r *Results) values() []string {
	return []string{
		r.Capacity, r.CarNumber, r.Car, r.Class,
		r.Driver, r.FastestLap, r.Gap, r.Laps,
		r.LapTime, r.Position, r.RaceTime, r.Team,
	}
}

func resultsFields() []string {
	return []string{
		"Capacity", "Car Number", "Car", "Class",
		"Driver", "Fastest Lap", "Gap", "Laps",
		"Lap Time", "Position", "Race Time", "Team",
	}
}
