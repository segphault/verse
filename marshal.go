package main

import (
	"encoding/json"
	"regexp"
)

type Regex struct {
	*regexp.Regexp
}

func (r *Regex) UnmarshalJSON(b []byte) error {
	str := new(string)
	json.Unmarshal(b, str)

	compiled, err := regexp.Compile(*str)

	if err != nil {
		return err
	}

	r.Regexp = compiled
	return nil
}
