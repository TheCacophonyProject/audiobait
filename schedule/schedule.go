package schedule

import "encoding/json"


type Schedule struct {
	Description string
	ControlNights int
	PlayNights int
	Combos []Combo
}

type Combo struct{
	From TimeOfDay
	Every int
	Until TimeOfDay
	Waits []int
	Volumes []int
	Sounds []string
}

func ParseJSONConfigFile(jsonAsString string, schedule *Schedule) error {
	data := []byte(jsonAsString);

	err := json.Unmarshal(data, &schedule)
	return err
}



