package main

import "fmt"

type Test struct {
	Value string
}

func (t Test) GetValue1() string {
	return t.Value
}
func (t *Test) GetValue2() string {
	return t.Value
}

func main() {
	t := Test{"fuck"}
	fmt.Println(t, t.GetValue1(), t.GetValue2())

	t2 := &t
	fmt.Println(t2, t2.GetValue1(), t2.GetValue2())

	t.Value = "FUCK"
	fmt.Println(t, t.GetValue1(), t.GetValue2())

	fmt.Println(t2, t2.GetValue1(), t2.GetValue2())
}
