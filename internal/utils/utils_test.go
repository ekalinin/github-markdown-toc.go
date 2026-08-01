package utils

import "testing"

func Test_RemoveStuff(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"All", "\n\nsome<code> code </code> here\n", "some code  here"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RemoveStuff(tt.in)
			if got != tt.want {
				t.Errorf("Got=%s, want=%s\n", got, tt.want)
			}
		})
	}
}
