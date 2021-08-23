package impart

import (
	"encoding/json"
)

type FilterEnum int

const (
	Gender_Man                        FilterEnum = 17
	Gender_NonBinary                  FilterEnum = 18
	Gender_NotListed                  FilterEnum = 19
	Gender_Woman                      FilterEnum = 16
	Generation_Boomer                 FilterEnum = 15
	Generation_GenX                   FilterEnum = 14
	Generation_GenZ                   FilterEnum = 12
	Generation_Millennial             FilterEnum = 13
	Household_Married                 FilterEnum = 4
	Household_Partner                 FilterEnum = 3
	Household_SharedCustody           FilterEnum = 5
	Household_Single                  FilterEnum = 1
	Household_SingleRoommates         FilterEnum = 2
	Race_AmIndianAlNative             FilterEnum = 20
	Race_AsianPacIslander             FilterEnum = 21
	Race_Black                        FilterEnum = 22
	Race_Hispanic                     FilterEnum = 23
	Race_SWAsianNAfrican              FilterEnum = 24
	Race_White                        FilterEnum = 25
	Dependents_None                   FilterEnum = 6
	Dependents_Other                  FilterEnum = 11
	Dependents_Parents                FilterEnum = 10
	Dependents_PostSchool             FilterEnum = 9
	Dependents_PreSchool              FilterEnum = 7
	Dependents_SchoolAge              FilterEnum = 8
	FinancialGoals_GenerationalWealth FilterEnum = 30
	FinancialGoals_House              FilterEnum = 28
	FinancialGoals_Philanthropy       FilterEnum = 29
	FinancialGoals_Retirement         FilterEnum = 26
	FinancialGoals_SaveCollege        FilterEnum = 27
)

func FilterData() ([]byte, error) {
	out, err := json.Marshal(map[string]FilterEnum{
		"Gender-Man":                        Gender_Man,
		"Gender-NonBinary":                  Gender_NonBinary,
		"Gender-Woman":                      Gender_Woman,
		"Gender-NotListed":                  Gender_NotListed,
		"Generation-Boomer":                 Generation_Boomer,
		"Generation-GenX":                   Generation_GenX,
		"Generation-GenZ":                   Generation_GenZ,
		"Generation-Millennial":             Generation_Millennial,
		"Household-Married":                 Household_Married,
		"Household-Partner":                 Household_Partner,
		"Household-SharedCustody":           Household_SharedCustody,
		"Household-Single":                  Household_Single,
		"Household-SingleRoommates":         Household_SingleRoommates,
		"Race-AmIndianAlNative":             Race_AmIndianAlNative,
		"Race-AsianPacIslander":             Race_AsianPacIslander,
		"Race-Black":                        Race_Black,
		"Race-Hispanic":                     Race_Hispanic,
		"Race-SWAsianNAfrican":              Race_SWAsianNAfrican,
		"Race-White":                        Race_White,
		"Dependents-None":                   Dependents_None,
		"Dependents-Other":                  Dependents_Other,
		"Dependents-Parents":                Dependents_Parents,
		"Dependents-PostSchool":             Dependents_PostSchool,
		"Dependents-PreSchool":              Dependents_PreSchool,
		"Dependents-SchoolAge":              Dependents_SchoolAge,
		"FinancialGoals-GenerationalWealth": FinancialGoals_GenerationalWealth,
		"FinancialGoals-House":              FinancialGoals_House,
		"FinancialGoals-Philanthropy":       FinancialGoals_Philanthropy,
		"FinancialGoals-Retirement":         FinancialGoals_Retirement,
		"FinancialGoals-SaveCollege":        FinancialGoals_SaveCollege,
	})
	if err != nil {
		return nil, err

	}
	return out, nil
}
