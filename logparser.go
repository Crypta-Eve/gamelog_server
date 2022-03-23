package main

import (
	"github.com/k3a/html2text"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	dateFormat = "2006.01.02 15:04:05"
)

var (
	miningMinedAmount   = regexp.MustCompile("mined [0-9]+ units")
	miningOreIdentifer  = regexp.MustCompile("of [A-Za-z ]* with")
	miningResidueAmount = regexp.MustCompile("residue of [0-9]+ units")
)

type (
	LogList struct {
		Name      string
		Character string
		Lines     []GameLog
	}

	GameLog struct {
		Timestamp time.Time
		Category  string
		Message   string
		Minings   []MiningAmount
	}

	MiningAmount struct {
		Amount int
		Ore    string
		Waste  int
	}
)

func parseLogLine(ll *LogList, line string) {
	gl := GameLog{}
	ts := line[2:21]
	t, err := time.Parse(dateFormat, ts)
	if err != nil {
		log.Println(err)
		return
	}
	gl.Timestamp = t

	cat := strings.Builder{}

	i := 25
	for i = 25; i < len(line); i++ {
		if line[i] == uint8(')') {
			break
		}
		cat.WriteRune(rune(line[i]))
	}
	gl.Category = cat.String()
	gl.Message = html2text.HTML2Text(line[i+1:])

	switch gl.Category {
	case "mining":
		parseMiningString(&gl)
	}

	if ll.Lines == nil {
		ll.Lines = []GameLog{}
	}
	ll.Lines = append(ll.Lines, gl)
}

func parseMiningString(gl *GameLog) {

	log.Println(gl.Message)

	// example string "You mined 21 units of Magma Mercoxit with a lost residue of 0 units"
	mined := miningMinedAmount.FindString(gl.Message)
	ore := miningOreIdentifer.FindString(gl.Message)
	residue := miningResidueAmount.FindString(gl.Message)

	log.Println(mined, ore, residue)

	mining := MiningAmount{
		Amount: 0,
		Ore:    "",
		Waste:  0,
	}
	if len(mined) > 6 {
		mining.Amount, _ = strconv.Atoi(strings.TrimSuffix(mined[6:], " units"))
	}
	if len(ore) > 3 {
		mining.Ore = strings.TrimSuffix(ore[3:], " with")
	}
	if len(residue) > 11 {
		mining.Waste, _ = strconv.Atoi(strings.TrimSuffix(residue[11:], " units"))
	}

	if mining.Amount > 0 || mining.Waste > 0 {
		if gl.Minings == nil {
			gl.Minings = []MiningAmount{}
		}
		gl.Minings = append(gl.Minings, mining)
	}
}
