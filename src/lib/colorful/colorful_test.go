package colorful

import "testing"

// TestFront 测试前景色
func TestFront(t *testing.T) {
	type arg struct {
		text  string
		color int
		args  []interface{}
	}
	var tests = []struct {
		in       arg    // input
		expected string // expected result
	}{
		{arg{"red", FrontRed, []interface{}{}}, "\033[1;31mred\033[0m"},
		{arg{"green", FrontGreen, []interface{}{}}, "\033[1;32mgreen\033[0m"},
		{arg{"yellow", FrontYellow, []interface{}{}}, "\033[1;33myellow\033[0m"},
		{arg{"blue", FrontBlue, []interface{}{}}, "\033[1;34mblue\033[0m"},
		{arg{"purple", FrontPurple, []interface{}{}}, "\033[1;35mpurple\033[0m"},
		{arg{"sky", FrontSky, []interface{}{}}, "\033[1;36msky\033[0m"},
		{arg{"red %d %d", R, []interface{}{1, 2}}, "\033[1;31mred 1 2\033[0m"},
	}

	for _, test := range tests {
		actual := Front(test.in.text, test.in.color, test.in.args...)
		if actual != test.expected {
			t.Errorf("[×] in: %v out: %q expected: %q\n", test.in, actual, test.expected)
		} else {
			t.Logf("[√] in: %v out: %v expected: %v\n", test.in, actual, test.expected)
		}
	}
}

// TestFront 测试背景色
func TestBack(t *testing.T) {
	type arg struct {
		text  string
		color int
	}
	var tests = []struct {
		in       arg    // input
		expected string // expected result
	}{
		{arg{"red", BackRed}, "\033[1;;41mred\033[0m"},
		{arg{"green", BackGreen}, "\033[1;;42mgreen\033[0m"},
		{arg{"yellow", BackYellow}, "\033[1;;43myellow\033[0m"},
		{arg{"blue", BackBlue}, "\033[1;;44mblue\033[0m"},
		{arg{"purple", BackPurple}, "\033[1;;45mpurple\033[0m"},
		{arg{"sky", BackSky}, "\033[1;;46msky\033[0m"},
	}

	for _, test := range tests {
		actual := Back(test.in.text, test.in.color)
		if actual != test.expected {
			t.Errorf("[×] in: %v out: %q expected: %q\n", test.in, actual, test.expected)
		} else {
			t.Logf("[√] in: %v out: %v expected: %v\n", test.in, actual, test.expected)
		}
	}
}
