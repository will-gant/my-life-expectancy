package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/araddon/dateparse"
	"github.com/iand/gedcom"
)

type DeathStat struct {
	Year             string
	LifeExpectancy   float64
	MedianAgeAtDeath float64
	ModalAgeAtDeath  float64
}

type AncestorDeath struct {
	Year                     int
	GenerationsRemoved       int
	Gender                   string
	AgeAtDeathDaysTotal      int
	AgeAtDeathYears          int
	AgeAtDeathDays           int
	LifeExpectancyDiffDays   int
	MedianAgeAtDeathDiffDays int
	ModalAgeAtDeathDiffDays  int
}

func calculateWeightedAverages(ancestors []AncestorDeath, gender string) (int, int, int) {
	var totalLifeExpectancyDiffDays, totalMedianAgeAtDeathDiffDays, totalModalAgeAtDeathDiffDays int
	var weightSum int

	highestGeneration := 0
	for _, ancestor := range ancestors {
		if ancestor.GenerationsRemoved > highestGeneration {
			highestGeneration = ancestor.GenerationsRemoved
		}
	}

	for _, ancestor := range ancestors {
		if ancestor.Gender == gender || gender == "" {
			weight := int(math.Pow(2, float64(highestGeneration-ancestor.GenerationsRemoved)))
			totalLifeExpectancyDiffDays += ancestor.LifeExpectancyDiffDays * weight
			totalMedianAgeAtDeathDiffDays += ancestor.MedianAgeAtDeathDiffDays * weight
			totalModalAgeAtDeathDiffDays += ancestor.ModalAgeAtDeathDiffDays * weight
			weightSum += weight
		}
	}

	return totalLifeExpectancyDiffDays / weightSum, totalMedianAgeAtDeathDiffDays / weightSum, totalModalAgeAtDeathDiffDays / weightSum
}

func checkValidYear(dateStr string) error {
	currentYear := time.Now().Year()
	re := regexp.MustCompile(`\b\d{4}\b`)
	yearStr := re.FindString(dateStr)
	if yearStr == "" {
		return fmt.Errorf("date '%s' does not contain a four-digit year", dateStr)
	}
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		return fmt.Errorf("error converting year from date '%s' to an int", dateStr)
	}
	if year < 1500 || year > currentYear {
		return fmt.Errorf("year in date '%s' is outside valid range", dateStr)
	}
	return nil
}

func parseDate(dateStr string) (time.Time, error) {
	err := checkValidYear(dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("no valid year found in date: %s", err)
	}

	dateStr = strings.TrimSpace(dateStr)
	if strings.Contains(dateStr, "Unknown") {
		return time.Time{}, fmt.Errorf("date is \"%s\"", dateStr)
	}
	dateStr = strings.TrimSpace(dateStr)
	dateStr = strings.ReplaceAll(dateStr, "  ", " ")

	months := []string{"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"}
	for _, month := range months {
		monthAbbr := month[:3]

		regex := regexp.MustCompile(fmt.Sprintf(`(?i)\b%s\b`, monthAbbr))
		dateStr = regex.ReplaceAllString(dateStr, month)

		if monthAbbr == "Sep" {
			monthAbbr = month[:4]
			regex := regexp.MustCompile(fmt.Sprintf(`(?i)\b%s\b`, monthAbbr))
			dateStr = regex.ReplaceAllString(dateStr, month)
		}
	}

	foundMonth := false
	for _, month := range months {
		if strings.HasPrefix(dateStr, month) {
			foundMonth = true
			break
		}
	}

	if !foundMonth {
		for i, r := range dateStr {
			if r >= '0' && r <= '9' {
				dateStr = dateStr[i:]
				break
			}
		}
	}

	yearRangeRegex := regexp.MustCompile(`(\d{4})\s*-\s*(\d{4})`)
	if yearRangeRegex.MatchString(dateStr) {
		dateStr = strings.ReplaceAll(dateStr, " ", "")
		parts := strings.Split(dateStr, "-")
		if len(parts) != 2 {
			return time.Time{}, fmt.Errorf("invalid date range: %s", dateStr)
		}
		startYear, err := strconv.Atoi(parts[0])
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid start year in date range: %s", dateStr)
		}
		endYear, err := strconv.Atoi(parts[1])
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid end year in date range: %s", dateStr)
		}
		start := time.Date(startYear, 1, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(endYear, 1, 1, 0, 0, 0, 0, time.UTC)
		midpoint := start.Add(end.Sub(start) / 2)
		return midpoint, nil
	}

	dateSuffixRegex := regexp.MustCompile(`(1|2)(st|nd|th)`)
	dateStr = dateSuffixRegex.ReplaceAllString(dateStr, "$1")

	for _, month := range months {
		if strings.Contains(dateStr, month) && !regexp.MustCompile(`\b\d{1,2} `+month).MatchString(dateStr) {
			dateStr = regexp.MustCompile(month).ReplaceAllString(dateStr, "1 "+month)
		}
	}

	parsedDate, err := dateparse.ParseLocal(dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse date %s: %s", dateStr, err)
	}

	return parsedDate, nil
}

func getDeathStatsForAncestors(ancestors map[*gedcom.IndividualRecord]int, maleDeathStats []DeathStat, femaleDeathStats []DeathStat) []AncestorDeath {
	var ancestorDeaths []AncestorDeath
	for individual, generation := range ancestors {
		var birthDate, deathDate time.Time
		hasBirth, hasDeath := true, true
		for _, event := range individual.Event {
			switch event.Tag {
			case "DEAT":
				var err error
				deathDate, err = parseDate(event.Date)
				if err != nil {
					hasDeath = false
				}
			case "BIRT":
				var err error
				birthDate, err = parseDate(event.Date)
				if err != nil {
					hasBirth = false
				}
			default:
				continue
			}
		}
		if !hasBirth || !hasDeath || birthDate == (time.Time{}) || deathDate == (time.Time{}) {
			continue
		}

		ageAtDeathDaysTotal := int(deathDate.Unix()-birthDate.Unix()) / 60 / 60 / 24

		var deathStats []DeathStat
		if strings.ToLower(individual.Sex) == "m" {
			deathStats = maleDeathStats
		} else {
			deathStats = femaleDeathStats
		}
		earliestYear, err := earliestYear(deathStats)
		if err != nil {
			fmt.Printf("unable to find earliest year in death stats: %s\n", err)
		}
		if deathDate.Year() < earliestYear {
			continue
		}
		var deathStat DeathStat
		for _, ds := range deathStats {
			if ds.Year == strconv.Itoa(deathDate.Year()) {
				deathStat = ds
				break
			}
		}
		LifeExpectancyDays := int(deathStat.LifeExpectancy * 365)
		MedianAgeAtDeathDays := int(deathStat.MedianAgeAtDeath * 365)
		ModalAgeAtDeathDays := int(deathStat.ModalAgeAtDeath * 365)
		ageAtDeathYears, ageAtDeathDays := daysToYearsAndDays(ageAtDeathDaysTotal)

		ancestorDeaths = append(ancestorDeaths, AncestorDeath{
			Year:                     deathDate.Year(),
			GenerationsRemoved:       generation,
			Gender:                   strings.ToLower(individual.Sex),
			AgeAtDeathYears:          ageAtDeathYears,
			AgeAtDeathDays:           ageAtDeathDays,
			AgeAtDeathDaysTotal:      ageAtDeathDaysTotal,
			LifeExpectancyDiffDays:   ageAtDeathDaysTotal - LifeExpectancyDays,
			MedianAgeAtDeathDiffDays: ageAtDeathDaysTotal - MedianAgeAtDeathDays,
			ModalAgeAtDeathDiffDays:  ageAtDeathDaysTotal - ModalAgeAtDeathDays,
		})
	}
	return ancestorDeaths
}

func parseDeathStats(filepath string) ([]DeathStat, error) {
	file, err := os.Open(filepath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)

	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error reading records:", err)
		return nil, err
	}

	var DeathStats []DeathStat

	for i, record := range records {
		if i == 0 {

			continue
		}

		var s DeathStat
		s.Year = record[0]
		s.LifeExpectancy, err = strconv.ParseFloat(record[1], 64)
		if err != nil {
			fmt.Println("Error parsing life expectancy:", err)
			return nil, err
		}

		s.MedianAgeAtDeath, err = strconv.ParseFloat(record[2], 64)
		if err != nil {
			fmt.Println("Error parsing median age at death:", err)
			return nil, err
		}

		s.ModalAgeAtDeath, err = strconv.ParseFloat(record[3], 64)
		if err != nil {
			fmt.Println("Error parsing modal age at death:", err)
			return nil, err
		}

		DeathStats = append(DeathStats, s)
	}
	return DeathStats, nil
}

func earliestYear(stats []DeathStat) (int, error) {
	var earliest int
	for _, s := range stats {
		year, err := strconv.Atoi(s.Year)
		if err != nil {
			fmt.Println("Error parsing year:", err)
			return 0, err
		}
		if earliest == 0 || year < earliest {
			earliest = year
		}
	}
	return earliest, nil
}

func getParents(individual *gedcom.IndividualRecord) ([]*gedcom.IndividualRecord, error) {
	var parents []*gedcom.IndividualRecord

	for _, parentRecord := range individual.Parents {
		if parentRecord.Family.Husband != nil {
			parents = append(parents, parentRecord.Family.Husband)
		}

		if parentRecord.Family.Wife != nil {
			parents = append(parents, parentRecord.Family.Wife)
		}
	}
	return parents, nil
}

func daysToYearsAndDays(daysTotal int) (int, int) {
	years := daysTotal / 365
	days := int(math.Abs(float64(daysTotal % 365)))
	return years, days
}

func printResults(ancestors []AncestorDeath) {
	_, maleTotalMedianAgeAtDeathDiffDays, maleTotalModalAgeAtDeathDiffDays := calculateWeightedAverages(ancestors, "m")
	_, femaleTotalMedianAgeAtDeathDiffDays, femaleTotalModalAgeAtDeathDiffDays := calculateWeightedAverages(ancestors, "f")
	_, overallTotalMedianAgeAtDeathDiffDays, overallTotalModalAgeAtDeathDiffDays := calculateWeightedAverages(ancestors, "")

	maleMedianAgeAtDeathYears, maleMedianAgeAtDeathDays := daysToYearsAndDays(maleTotalMedianAgeAtDeathDiffDays)
	maleModalAgeAtDeathYears, maleModalAgeAtDeathDays := daysToYearsAndDays(maleTotalModalAgeAtDeathDiffDays)

	femaleMedianAgeAtDeathYears, femaleMedianAgeAtDeathDays := daysToYearsAndDays(femaleTotalMedianAgeAtDeathDiffDays)
	femaleModalAgeAtDeathYears, femaleModalAgeAtDeathDays := daysToYearsAndDays(femaleTotalModalAgeAtDeathDiffDays)

	overallMedianAgeAtDeathYears, overallMedianAgeAtDeathDays := daysToYearsAndDays(overallTotalMedianAgeAtDeathDiffDays)
	overallModalAgeAtDeathYears, overallModalAgeAtDeathDays := daysToYearsAndDays(overallTotalModalAgeAtDeathDiffDays)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "===========================================================================================")
	fmt.Fprintln(w, "Stat\tMale\tFemale\tOverall")
	fmt.Fprintln(w, "Difference from Median Death Age\t"+strconv.Itoa(maleMedianAgeAtDeathYears)+" "+"years"+" "+strconv.Itoa(maleMedianAgeAtDeathDays)+" "+"days\t"+strconv.Itoa(femaleMedianAgeAtDeathYears)+" "+"years"+" "+strconv.Itoa(femaleMedianAgeAtDeathDays)+" "+"days\t"+strconv.Itoa(overallMedianAgeAtDeathYears)+" "+"years"+" "+strconv.Itoa(overallMedianAgeAtDeathDays)+" "+"days")
	fmt.Fprintln(w, "Difference from Modal Age at Death\t"+strconv.Itoa(maleModalAgeAtDeathYears)+" "+"years"+" "+strconv.Itoa(maleModalAgeAtDeathDays)+" "+"days\t"+strconv.Itoa(femaleModalAgeAtDeathYears)+" "+"years"+" "+strconv.Itoa(femaleModalAgeAtDeathDays)+" "+"days\t"+strconv.Itoa(overallModalAgeAtDeathYears)+" "+"years"+" "+strconv.Itoa(overallModalAgeAtDeathDays)+" "+"days")
	w.Flush()
	fmt.Fprintln(w, "===========================================================================================")
}

func getAncestors(individual *gedcom.IndividualRecord, ancestors map[*gedcom.IndividualRecord]int, generation int) (map[*gedcom.IndividualRecord]int, error) {
	parents, err := getParents(individual)
	if err != nil {
		return nil, err
	}

	for _, parent := range parents {
		ancestors[parent] = generation
	}

	for _, parent := range parents {
		ancestors, err = getAncestors(parent, ancestors, generation+1)
		if err != nil {
			return nil, err
		}
	}

	return ancestors, nil
}

func main() {
	var treeFile string
	flag.StringVar(&treeFile, "tree-file", "", "path to GEDCOM tree file")
	flag.Parse()
	if treeFile == "" {
		fmt.Println("Error: --tree-file flag is required")
		os.Exit(1)
	}
	maleDeathStats, err := parseDeathStats("male_death_stats.csv")
	if err != nil {
		fmt.Printf("Error parsing male death stats: %v", err)
		os.Exit(1)
	}
	femaleDeathStats, err := parseDeathStats("female_death_stats.csv")
	if err != nil {
		fmt.Printf("Error parsing female death stats: %v", err)
		os.Exit(1)
	}

	data, _ := ioutil.ReadFile(treeFile)
	d := gedcom.NewDecoder(bytes.NewReader(data))
	g, _ := d.Decode()
	subject := g.Individual[0]

	ancestors, err := getAncestors(subject, map[*gedcom.IndividualRecord]int{}, 1)
	if err != nil {
		fmt.Printf("Error retrieving direct ancestors: %v", err)
		os.Exit(1)
	}

	ancestorDeaths := getDeathStatsForAncestors(ancestors, maleDeathStats, femaleDeathStats)
	printResults(ancestorDeaths)
}
