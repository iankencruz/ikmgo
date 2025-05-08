package utils

import "net/url"

type Form struct {
	Values         url.Values
	Errors         map[string]string // Field-specific
	NonFieldErrors []string          // General errors
}

func NewForm(data url.Values) *Form {
	return &Form{
		Values:         data,
		Errors:         make(map[string]string),
		NonFieldErrors: []string{},
	}
}

func (f *Form) Required(fields ...string) {
	for _, field := range fields {
		value := f.Values.Get(field)
		if value == "" {
			f.Errors[field] = "This field is required"
		}
	}
}

func (f *Form) Valid() bool {
	return len(f.Errors) == 0 && len(f.NonFieldErrors) == 0
}
