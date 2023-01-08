package main

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/iand/gedcom"
)

func TestParseDeathStats(t *testing.T) {
	stats, err := parseDeathStats("male_death_stats.csv")
	if err != nil {
		t.Errorf("Error parsing death stats: %v", err)
	}
	if len(stats) == 0 {
		t.Error("Expected at least one death stat, got none")
	}
}

func TestCheckValidYear(t *testing.T) {
	tests := []struct {
		dateStr  string
		expected error
	}{
		{dateStr: "1 Jan 2010", expected: nil},
		{dateStr: "1 Jan a year", expected: fmt.Errorf("date '1 Jan a year' does not contain a four-digit year")},
		{dateStr: "1 Jan 9999", expected: fmt.Errorf("year in date '1 Jan 9999' is outside valid range")},
	}

	for _, test := range tests {
		err := checkValidYear(test.dateStr)
		if err != nil {
			if err.Error() != test.expected.Error() {
				t.Errorf("unexpected error for dateStr %q: got %q, want %q", test.dateStr, err, test.expected)
			}
		} else if test.expected != nil {
			t.Errorf("did not get expected error for dateStr %q: got nil, want %q", test.dateStr, test.expected)
		}
	}
}

func TestGetDeathStatsForAncestors(t *testing.T) {
	maleDeathStats := []DeathStat{
		{Year: "1900", LifeExpectancy: 40.0, MedianAgeAtDeath: 40.0, ModalAgeAtDeath: 40.0},
		{Year: "1950", LifeExpectancy: 49.3, MedianAgeAtDeath: 46.5, ModalAgeAtDeath: 43.8},
	}

	femaleDeathStats := []DeathStat{
		{Year: "1800", LifeExpectancy: 50.5, MedianAgeAtDeath: 48.3, ModalAgeAtDeath: 45.6},
		{Year: "1850", LifeExpectancy: 50.9, MedianAgeAtDeath: 48.5, ModalAgeAtDeath: 45.7},
	}

	validAncestor := &gedcom.IndividualRecord{
		Sex: "m",
		Event: []*gedcom.EventRecord{
			{
				Tag:  "BIRT",
				Date: "1 Jan 1860",
			},
			{
				Tag:  "DEAT",
				Date: "1 Jan 1900",
			},
		},
	}

	invalidAncestor1 := &gedcom.IndividualRecord{
		Sex: "f",
		Event: []*gedcom.EventRecord{
			{
				Tag:  "BIRT",
				Date: "1 Jan 1922",
			},
		},
	}

	invalidAncestor2 := &gedcom.IndividualRecord{
		Sex: "m",
		Event: []*gedcom.EventRecord{
			{
				Tag:  "DEAT",
				Date: "1 Jan 1950",
			},
		},
	}

	invalidAncestor3 := &gedcom.IndividualRecord{
		Sex: "f",
		Event: []*gedcom.EventRecord{
			{
				Tag:  "BIRT",
				Date: "1 Jan 1900",
			},
			{
				Tag:  "DEAT",
				Date: "not a date",
			},
		},
	}

	ancestors := map[*gedcom.IndividualRecord]int{
		validAncestor:    1,
		invalidAncestor1: 1,
		invalidAncestor2: 1,
		invalidAncestor3: 1,
	}

	tests := []struct {
		name             string
		ancestors        map[*gedcom.IndividualRecord]int
		maleDeathStats   []DeathStat
		femaleDeathStats []DeathStat
		want             []AncestorDeath
	}{
		{
			name:             "valid ancestor",
			ancestors:        ancestors,
			maleDeathStats:   maleDeathStats,
			femaleDeathStats: femaleDeathStats,
			want: []AncestorDeath{
				{
					Year:                     1900,
					GenerationsRemoved:       1,
					Gender:                   "m",
					AgeAtDeathYears:          40,
					AgeAtDeathDays:           10,
					AgeAtDeathDaysTotal:      14610,
					LifeExpectancyDiffDays:   10,
					MedianAgeAtDeathDiffDays: 10,
					ModalAgeAtDeathDiffDays:  10,
				},
			},
		},
	}

	for _, test := range tests {
		got := getDeathStatsForAncestors(test.ancestors, test.maleDeathStats, test.femaleDeathStats)
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("test %q: got %v, want %v", test.name, got, test.want)
		}
	}

}

func TestParseDate(t *testing.T) {
	testCases := map[string]time.Time{
		"2009":             time.Date(2009, 1, 1, 0, 0, 0, 0, time.Local),
		"15 February 1943": time.Date(1943, 2, 15, 0, 0, 0, 0, time.Local),
		"15 Feb 1943":      time.Date(1943, 2, 15, 0, 0, 0, 0, time.Local),
		"Sep 1984":         time.Date(1984, 9, 1, 0, 0, 0, 0, time.Local),
		"Sept 1984":        time.Date(1984, 9, 1, 0, 0, 0, 0, time.Local),
		"2 Sept 1984":      time.Date(1984, 9, 2, 0, 0, 0, 0, time.Local),
		"2005-2007":        time.Date(2006, 1, 1, 0, 0, 0, 0, time.UTC),
		"2004 - 2007":      time.Date(2005, 7, 2, 0, 0, 0, 0, time.UTC),
		"About 1800":       time.Date(1800, 1, 1, 0, 0, 0, 0, time.Local),
		"about 1800":       time.Date(1800, 1, 1, 0, 0, 0, 0, time.Local),
		"abt 1800":         time.Date(1800, 1, 1, 0, 0, 0, 0, time.Local),
		"abt. 1800":        time.Date(1800, 1, 1, 0, 0, 0, 0, time.Local),
		"15  March  1900":  time.Date(1900, 3, 15, 0, 0, 0, 0, time.Local),
		"21st June 1850":   time.Date(1850, 6, 21, 0, 0, 0, 0, time.Local),
		"05/12/1851":       time.Date(1851, 5, 12, 0, 0, 0, 0, time.Local),
		"5/12/1851":        time.Date(1851, 5, 12, 0, 0, 0, 0, time.Local),
	}

	for dateStr, expectedParsedDate := range testCases {
		t.Run(dateStr, func(t *testing.T) {
			parsed, err := parseDate(dateStr)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if !parsed.Equal(expectedParsedDate) {
				t.Errorf("expected %s to parse as %s but got %s", dateStr, expectedParsedDate, parsed)
			}
		})
	}
}

func TestEarliestYear(t *testing.T) {
	stats := []DeathStat{
		{Year: "2000"},
		{Year: "2010"},
		{Year: "2020"},
	}
	year, err := earliestYear(stats)
	if err != nil {
		t.Errorf("Error getting earliest year: %v", err)
	}
	if year != 2000 {
		t.Errorf("Expected earliest year to be 2000, got %d", year)
	}
}

func TestGetParents(t *testing.T) {
	individual := &gedcom.IndividualRecord{
		Parents: []*gedcom.FamilyLinkRecord{
			{
				Family: &gedcom.FamilyRecord{
					Husband: &gedcom.IndividualRecord{},
					Wife:    &gedcom.IndividualRecord{},
				},
			},
		},
	}
	father := individual.Parents[0].Family.Husband
	mother := individual.Parents[0].Family.Wife

	parents, err := getParents(individual)
	if err != nil {
		t.Errorf("Error getting parents: %v", err)
	}

	if len(parents) != 2 {
		t.Errorf("Expected 2 parents, got %d", len(parents))
	}

	if !containsIndividual(parents, father) {
		t.Errorf("Expected father to be in parent list")
	}
	if !containsIndividual(parents, mother) {
		t.Errorf("Expected mother to be in parent list")
	}
}

func TestGetParentsMultipleFamilyLinkRecords(t *testing.T) {
	individual := &gedcom.IndividualRecord{
		Parents: []*gedcom.FamilyLinkRecord{
			{
				Family: &gedcom.FamilyRecord{
					Husband: &gedcom.IndividualRecord{},
				},
			},
			{
				Family: &gedcom.FamilyRecord{
					Wife: &gedcom.IndividualRecord{},
				},
			},
		},
	}
	father := individual.Parents[0].Family.Husband
	mother := individual.Parents[1].Family.Wife

	parents, err := getParents(individual)
	if err != nil {
		t.Errorf("Error getting parents: %v", err)
	}

	if len(parents) != 2 {
		t.Errorf("Expected 2 parents, got %d", len(parents))
	}

	if !containsIndividual(parents, father) {
		t.Errorf("Expected father to be in parent list")
	}

	if !containsIndividual(parents, mother) {
		t.Errorf("Expected mother to be in parent list")
	}
}

func containsIndividual(individuals []*gedcom.IndividualRecord, individual *gedcom.IndividualRecord) bool {
	for _, i := range individuals {
		if i == individual {
			return true
		}
	}
	return false
}
func TestGetAncestors(t *testing.T) {
	subject := &gedcom.IndividualRecord{
		Parents: []*gedcom.FamilyLinkRecord{
			{
				Family: &gedcom.FamilyRecord{
					Husband: &gedcom.IndividualRecord{
						Parents: []*gedcom.FamilyLinkRecord{
							{
								Family: &gedcom.FamilyRecord{
									Husband: &gedcom.IndividualRecord{},
									Wife:    &gedcom.IndividualRecord{},
								},
							},
						},
					},
					Wife: &gedcom.IndividualRecord{
						Parents: []*gedcom.FamilyLinkRecord{
							{
								Family: &gedcom.FamilyRecord{
									Husband: &gedcom.IndividualRecord{},
									Wife:    &gedcom.IndividualRecord{},
								},
							},
						},
					},
				},
			},
		},
	}
	father := subject.Parents[0].Family.Husband
	mother := subject.Parents[0].Family.Wife
	fathersMother := father.Parents[0].Family.Wife
	fathersFather := father.Parents[0].Family.Husband
	mothersMother := mother.Parents[0].Family.Wife
	mothersFather := mother.Parents[0].Family.Husband

	ancestors, err := getAncestors(subject, map[*gedcom.IndividualRecord]int{}, 1)
	if err != nil {
		t.Errorf("Error getting ancestors: %v", err)
	}

	if len(ancestors) != 6 {
		t.Errorf("Expected 6 ancestors, got %d", len(ancestors))
	}

	if ancestors[father] != 1 {
		t.Errorf("Expected father to be 1 generation removed, got %d", ancestors[father])
	}

	if ancestors[mother] != 1 {
		t.Errorf("Expected mother to be 1 generation removed, got %d", ancestors[mother])
	}

	if ancestors[fathersMother] != 2 {
		t.Errorf("Expected paternal grandmother to be 2 generation removed, got %d", ancestors[fathersMother])
	}

	if ancestors[fathersFather] != 2 {
		t.Errorf("Expected paternal grandfather to be 2 generation removed, got %d", ancestors[fathersFather])
	}

	if ancestors[mothersFather] != 2 {
		t.Errorf("Expected maternal grandfather to be 2 generation removed, got %d", ancestors[mothersFather])
	}

	if ancestors[mothersMother] != 2 {
		t.Errorf("Expected maternal grandmother to be 2 generation removed, got %d", ancestors[mothersMother])
	}
}

func TestCalculateWeightedAverages(t *testing.T) {
	ancestors := []AncestorDeath{
		{
			Gender:                   "f",
			LifeExpectancyDiffDays:   4,
			MedianAgeAtDeathDiffDays: 5,
			ModalAgeAtDeathDiffDays:  6,
			GenerationsRemoved:       1,
		},
		{
			Gender:                   "f",
			LifeExpectancyDiffDays:   10,
			MedianAgeAtDeathDiffDays: 11,
			ModalAgeAtDeathDiffDays:  12,
			GenerationsRemoved:       2,
		},
		{
			Gender:                   "m",
			LifeExpectancyDiffDays:   50,
			MedianAgeAtDeathDiffDays: 500,
			ModalAgeAtDeathDiffDays:  10,
			GenerationsRemoved:       1,
		},
		{
			Gender:                   "m",
			LifeExpectancyDiffDays:   100,
			MedianAgeAtDeathDiffDays: 1000,
			ModalAgeAtDeathDiffDays:  20,
			GenerationsRemoved:       3,
		},
	}

	tests := []struct {
		name       string
		ancestors  []AncestorDeath
		gender     string
		wantLife   int
		wantMedian int
		wantModal  int
	}{
		{
			name:       "female",
			ancestors:  ancestors,
			gender:     "f",
			wantLife:   6,
			wantMedian: 7,
			wantModal:  8,
		},
		{
			name:       "male",
			ancestors:  ancestors,
			gender:     "m",
			wantLife:   60,
			wantMedian: 600,
			wantModal:  12,
		},
		{
			name:       "all",
			ancestors:  ancestors,
			gender:     "",
			wantLife:   30,
			wantMedian: 276,
			wantModal:  9,
		},
	}

	for _, test := range tests {
		gotLife, gotMedian, gotModal := calculateWeightedAverages(test.ancestors, test.gender)
		if gotLife != test.wantLife {
			t.Errorf("test %q: got %v, want %v", test.name, gotLife, test.wantLife)
		}
		if gotMedian != test.wantMedian {
			t.Errorf("test %q: got %v, want %v", test.name, gotMedian, test.wantMedian)
		}
		if gotModal != test.wantModal {
			t.Errorf("test %q: got %v, want %v", test.name, gotModal, test.wantModal)
		}
	}
}
