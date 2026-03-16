package application

type UpdateSettingCommand struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
