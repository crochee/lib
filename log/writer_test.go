// Package log
package log

import (
	"io"
	"path/filepath"
	"reflect"
	"testing"
)

func TestSetLoggerWriter(t *testing.T) {
	path := filepath.Clean("./log/" + "op.log")
	t.Log(path)
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(abs)
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want io.Writer
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SetLoggerWriter(tt.args.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetLoggerWriter() = %v, want %v", got, tt.want)
			}
		})
	}
}
