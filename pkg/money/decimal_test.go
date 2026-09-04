package money

import "testing"

func TestParse(t *testing.T) {
	for _, tc := range []struct{ input, want string }{
		{"0", "0.00"}, {"-0.00", "0.00"}, {"2.9", "2.90"},
		{"-12.34", "-12.34"}, {"9007199254740993.01", "9007199254740993.01"},
	} {
		t.Run(tc.input, func(t *testing.T) {
			value, err := Parse(tc.input)
			if err != nil {
				t.Fatal(err)
			}
			if value.FloatString(2) != tc.want {
				t.Fatalf("got %s, want %s", value.FloatString(2), tc.want)
			}
		})
	}
	for _, input := range []string{"", "NaN", "Inf", "1/2", "1.234", "1.00oops", " 1.00", "1e3"} {
		t.Run("invalid/"+input, func(t *testing.T) {
			if _, err := Parse(input); err == nil {
				t.Fatal("accepted invalid money")
			}
		})
	}
}
