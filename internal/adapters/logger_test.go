package adapters

import "testing"

type fakeLogger struct {
	output string
}

func (l *fakeLogger) Info(format string, v ...any) {
	l.output = format
}

func Test_Logger(t *testing.T) {
	tests := []struct {
		name   string
		logger *Logger
		want   string
	}{
		{"With debug", NewLoggerX(true, &fakeLogger{}), "log it now"},
		{"No debug", NewLoggerX(false, &fakeLogger{}), ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.logger.Info("log it now")
			logger := tt.logger.log.(*fakeLogger)
			got := logger.output
			if got != tt.want {
				t.Errorf("Got=%s, want=%s", got, tt.want)
			}
		})
	}
}
