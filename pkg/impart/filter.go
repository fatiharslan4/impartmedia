package impart

import (
	"encoding/json"
)

type FilterEnum int

const (
	Gender_Man                            FilterEnum = 17
	Gender_NonBinary                                 = 18
	Gender_NotListed                                 = 19
	Gender_Woman                                     = 16
	Generation_Boomer                                = 15
	Generation_GenX                                  = 14
	Generation_GenZ                                  = 12
	Generation_Millennial                            = 13
	Household_Married                                = 4
	Household_Partner                                = 3
	Household_SharedCustody                          = 5
	Household_Single                                 = 1
	Household_SingleRoommates                        = 2
	Race_AmIndianAlNative                            = 20
	Race_AsianPacIslander                            = 21
	Race_Black                                       = 22
	Race_Hispanic                                    = 23
	Race_SWAsianNAfrican                             = 24
	Race_White                                       = 25
	Dependents_None                                  = 6
	Dependents_Other                                 = 11
	Dependents_Parents                               = 10
	Dependents_PostSchool                            = 9
	Dependents_PreSchool                             = 7
	Dependents_SchoolAge                             = 8
	FinancialGoals_GenerationalWealth                = 30
	FinancialGoals_House                             = 28
	FinancialGoals_Philanthropy                      = 29
	FinancialGoals_Retirement                        = 26
	FinancialGoals_SaveCollege                       = 27
	Industry_Agriculture                             = 31
	Industry_Business                                = 32
	Industry_Construction                            = 33
	Industry_Education                               = 34
	Industry_EntertainmentGaming                     = 35
	Industry_Financensurance                         = 36
	Industry_FoodHospitality                         = 37
	Industry_GovernmentPublicServices                = 38
	Industry_HealthServices                          = 39
	Industry_Legal                                   = 40
	Industry_NaturalResources                        = 41
	Industry_PersonalProfessionalServices            = 42
	Industry_RealEstateHousing                       = 43
	Industry_RetaileCommerce                         = 44
	Industry_SafetySecurity                          = 45
	Industry_Transportation                          = 46
	Career_Entrylevel                                = 47
	Career_Midlevel                                  = 47
	Career_Management                                = 49
	Career_UpperManagement                           = 50
	Career_BusinessOwner                             = 51
	Career_Other                                     = 52
	Income_Income0                                   = 53
	Income_Income1                                   = 54
	Income_Income2                                   = 55
	Income_Income3                                   = 56
	Income_Income4                                   = 57
	Income_Income5                                   = 58
	Income_Income6                                   = 66
	Income_Income7                                   = 67
	Income_Income8                                   = 68
	EmploymentStatus_FullTime                        = 59
	EmploymentStatus_PartTime                        = 60
	EmploymentStatus_Unemployed                      = 61
	EmploymentStatus_Self                            = 62
	EmploymentStatus_HomeMaker                       = 63
	EmploymentStatus_Student                         = 64
	EmploymentStatus_Retired                         = 65
)

func FilterData() ([]byte, error) {
	out, err := json.Marshal(map[string]FilterEnum{
		"Gender-Man":                                    Gender_Man,
		"Gender-Non-binary":                             Gender_NonBinary,
		"Gender-Woman":                                  Gender_Woman,
		"Gender-Not listed":                             Gender_NotListed,
		"Generation-Boomer (born 1946-1964)":            Generation_Boomer,
		"Generation-Gen X (born 1965-1980)":             Generation_GenX,
		"Generation-Gen Z (born after 2001)":            Generation_GenZ,
		"Generation-Millennial (born 1981-2000)":        Generation_Millennial,
		"Household-Married":                             Household_Married,
		"Household-Living with partner":                 Household_Partner,
		"Household-Shared custody":                      Household_SharedCustody,
		"Household-Single adult":                        Household_Single,
		"Household-Single living with others":           Household_SingleRoommates,
		"Race-American Indian/Alaskan Native":           Race_AmIndianAlNative,
		"Race-Asian/Pacific Islander":                   Race_AsianPacIslander,
		"Race-Black/African American":                   Race_Black,
		"Race-Hispanic/Latino":                          Race_Hispanic,
		"Race-Southwestern Asian/North African":         Race_SWAsianNAfrican,
		"Race-White":                                    Race_White,
		"Dependents-None":                               Dependents_None,
		"Dependents-Other family members":               Dependents_Other,
		"Dependents-Parents":                            Dependents_Parents,
		"Dependents-Post school children (19+)":         Dependents_PostSchool,
		"Dependents-Pre-school children (0-4)":          Dependents_PreSchool,
		"Dependents-School age children (5-18)":         Dependents_SchoolAge,
		"Financial Goals-Generational wealth or legacy": FinancialGoals_GenerationalWealth,
		"Financial Goals-House down payment":            FinancialGoals_House,
		"Financial Goals-Philanthropy":                  FinancialGoals_Philanthropy,
		"Financial Goals-Retirement":                    FinancialGoals_Retirement,
		"Financial Goals-Save for college":              FinancialGoals_SaveCollege,
		"Industry-Agriculture & Forestry/Wildlife":      Industry_Agriculture,
		"Industry-Business & Technology":                Industry_Business,
		"Industry-Construction/Utilities/Contracting":   Industry_Construction,
		"Industry-Education":                            Industry_Education,
		"Industry-Entertainment & Gaming":               Industry_EntertainmentGaming,
		"Industry-Finance & Insurance":                  Industry_Financensurance,
		"Industry-Food & Hospitality":                   Industry_FoodHospitality,
		"Industry-Government & Public Services":         Industry_GovernmentPublicServices,
		"Industry-Health Services & Healthcare":         Industry_HealthServices,
		"Industry-Legal":                                Industry_Legal,
		"Industry-Natural Resources/Environmental":      Industry_NaturalResources,
		"Industry-Personal & Professional Services":     Industry_PersonalProfessionalServices,
		"Industry-Real Estate & Housing":                Industry_RealEstateHousing,
		"Industry-Retail & eCommerce":                   Industry_RealEstateHousing,
		"Industry-Safety & Security":                    Industry_SafetySecurity,
		"Industry-Transportation":                       Industry_Transportation,
		"Career-Entry-level":                            Career_Entrylevel,
		"Career-Mid-level":                              Career_Midlevel,
		"Career-Management":                             Career_Management,
		"Career-Upper Management":                       Career_UpperManagement,
		"Career-Business Owner":                         Career_BusinessOwner,
		"Career-Other":                                  Career_Other,
		"Income-Less than $25,000":                      Income_Income0,
		"Income-$25,000 - $34,999":                      Income_Income1,
		"Income-$35,000 - $49,999":                      Income_Income2,
		"Income-$50,000 - $74,999":                      Income_Income3,
		"Income-$75,000 - $99,999":                      Income_Income4,
		"Income-$100,000 - $149,999":                    Income_Income5,
		"Income-$150,000 - $199,999":                    Income_Income6,
		"Income-$200,000 - $299,999":                    Income_Income7,
		"Income-More than $300,000":                     Income_Income8,
		"EmploymentStatus-Full-time employment":         EmploymentStatus_FullTime,
		"EmploymentStatus-Part-time employment":         EmploymentStatus_PartTime,
		"EmploymentStatus-Unemployed":                   EmploymentStatus_Unemployed,
		"EmploymentStatus-Self-employed":                EmploymentStatus_Self,
		"EmploymentStatus-Home-maker":                   EmploymentStatus_HomeMaker,
		"EmploymentStatus-Student":                      EmploymentStatus_Student,
		"EmploymentStatus-Retired":                      EmploymentStatus_Retired,
	})
	if err != nil {
		return nil, err

	}
	return out, nil
}

type CommonCheckEnum string

const (
	AddToWaitlist             = "addto_waitlist"
	AddToHive                 = "addto_hive"
	AddToAdmin                = "addto_admin"
	Hive_mail                 = "hive_email"
	Waitlist_mail             = "waitlist_email"
	Waitlist_mail_subject     = "You’re on the Hive Wealth waitlist!"
	Waitlist_mail_previewtext = "Thank you for signing up – we’ll let you know when we’ve found your Hive"
	Hive_mail_subject         = "Great news! We’ve found your Hive!"
	Hive_mail_previewtext     = "What now? Let the journey to your financial empowerment begin!"

)

const (
	IncludeAdmin = 1
	ExcludeAdmin = 0
	IncludeAll   = 2
)
