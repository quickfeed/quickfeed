import (
	"time"
	"strings"
)

_effort: {
	#min: >=1 & int
	#max: >#min & int
	"\(#min)-\(#max) hours"
}

// could use reg.exp to require capitalized first letter in heading/description/title etc (see the video on cuelang.org)
#criteria: {
	description: strings.MaxRunes(5)
	points:      >=1 & <=100 & int
}

#benchmark: {
	heading:  string
	criteria: #criteria
}

#Assignment: {
	order:            >=1 & int
	title:            string
	effort:           string
	deadline:         string & (time.Format("2006-1-2T15:04") | time.Format("2006-1-2 15:04"))
	containertimeout: string & (*"10m" | time.Format("4m") | time.Format("5s"))
	type:             *"individual" | "group"
	approve:          *"manual" | "auto"
	approveScore:     >=60 & <=100 & int
	if approve == "manual" {
		reviewers: *1 | 2 | 3 & int
		benchmarks?: [...#benchmark]
	}
}

lab1: #Assignment & {
	order:            1
	title:            "Introduction to Unix"
	type:             "group"
	deadline:         "2020-08-30 23:59"
	reviewers:        2
	approveScore:     90
	containertimeout: "30s"
	effort:           _effort & {_, #min: 5, #max: 10}
	benchmarks: [
		{heading: "hey", criteria: {description: "hoho", points: 5}},
	]
}

lab2: #Assignment & {
	order:        2
	title:        "Introduction to Unix 2"
	deadline:     "2020-09-30T23:59"
	type:         "individual"
	approve:      "auto"
	approveScore: 60
	effort:       _effort & {_, #min: 9, #max: 10}
}

lab3: (#Assignment & {
	order:            3
	title:            "Introduction to Unix 3"
	deadline:         "2020-10-30T23:59"
	approveScore:     60
	containertimeout: "2m"
	effort:           _effort & {_, #min: 5, #max: 10}
})
