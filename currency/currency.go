package currency

import "errors"

var currencies = map[string]string{
	"USD": "US Dollar",
	"EUR": "Euro",
	"CAD": "Canadian Dollar",
}

func IsSupportedCurrency(currency string) bool {

	// check if currency identifier is a key of the map
	if _, ok := currencies[currency]; ok {
		return true
	}

	return false
}

func GetFullCurrencyName(currency string) (string, error) {
	val, ok := currencies[currency]

	if ok {
		return val, nil
	}

	return "", errors.New("received unsupported currency identifier")
}

func GetSupportedCurrencies() []string {
	var output []string

	for k, _ := range currencies {
		output = append(output, k)
	}

	return output
}
