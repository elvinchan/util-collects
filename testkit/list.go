package testkit

import "testing"

type List []struct {
	Name, Got, Exp string
}

func (l List) Validate(t *testing.T) {
	for _, test := range l {
		if test.Got != test.Exp {
			t.Errorf("On %v, expected '%v', but got '%v'",
				test.Name, test.Exp, test.Got)
		}
	}
}
