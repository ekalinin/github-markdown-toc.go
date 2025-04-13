package adapters

import (
	"os"
	"testing"
)

func Test_FileWriter(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"Written"},
	}
	writer := NewFileWriter(NewLogger(false))
	checker := NewFileCheck(NewLogger(false))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := "./tmp-for-test"
			test_data := "some-test"
			err := writer.Write(file, []byte(test_data))
			if err != nil {
				t.Errorf("Got err=%v", err)
			}
			if !checker.Exists(file) {
				t.Errorf("File not exists, f=%v", file)
			}
			data, err := os.ReadFile(file)
			if err != nil {
				t.Errorf("Got read err=%v", err)
			}
			if got := string(data); got != test_data {
				t.Errorf("Got=%v, want=%v", got, test_data)
			}
			err = os.Remove(file)
			if err != nil {
				t.Errorf("Error on delete file=%v err=%v", file, err)
			}
		})
	}
}
