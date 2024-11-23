package notRemoved

import "slices"

var Data []string

func Init() {
	Data = append(Data, "foo")
	Data = append(Data, "bar")
	Data = append(Data, "baz")
}

func Remove(s string) {
	Data = slices.DeleteFunc(Data, func(v string) bool {
		return v == s
	})
}
