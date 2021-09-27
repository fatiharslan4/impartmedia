package impart

//  Use questionenum no from database table -Questions.
type QuestionEnum int

const (
	Gender           QuestionEnum = 4
	Household                     = 1
	Race                          = 5
	Dependents                    = 2
	FinancialGoals                = 6
	Industry                      = 7
	Career                        = 8
	Income                        = 9
	Generation                    = 3
	EmploymentStatus              = 10
)

func GetUserAnswerList() map[uint]string {
	userAnswer := make(map[uint]string)
	start := 1
	limit := 10 // based on the Question number (i.e check limit as the QuestionEnum based.Need to set empty for all fields in QuestionEnum)
	for start <= limit {
		userAnswer[uint(start)] = ""
		start = start + 1
	}
	return userAnswer
}

func SetMailChimpAnswer(userAnswer map[uint]string, status string, zipCode string) map[string]interface{} {
	data := make(map[string]interface{})
	data["STATUS"] = status
	data["GENDER"] = userAnswer[uint(Gender)]
	data["HOUSEHOLD"] = userAnswer[uint(Household)]
	data["DEPENDENTS"] = userAnswer[uint(Dependents)]
	data["GENERATION"] = userAnswer[uint(Generation)]
	data["RACE"] = userAnswer[uint(Race)]
	data["FINANCIALG"] = userAnswer[uint(FinancialGoals)]
	data["INDUSTRY"] = userAnswer[uint(Industry)]
	data["CAREER"] = userAnswer[uint(Career)]
	data["INCOME"] = userAnswer[uint(Income)]
	data["EMPLOYMENT"] = userAnswer[uint(EmploymentStatus)]
	data["ZIPCODE"] = zipCode
	return data
}
