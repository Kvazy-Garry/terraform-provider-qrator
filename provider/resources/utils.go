package resources

import "encoding/json"

// ToJSON — преобразование значения в строку JSON
func ToJSON(value interface{}) (string, error) {
	bytes, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// FromJSON — обратное преобразование строки JSON в интерфейс
func FromJSON(str string, result interface{}) error {
	return json.Unmarshal([]byte(str), result)
}
