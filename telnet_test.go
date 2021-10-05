package telnetssh

import (
	"reflect"
	"testing"

	"github.com/muxache/telnetssh/model"
)

func TestNewTelnetConnectionAndAuth(t *testing.T) {
	type args struct {
		c *model.Config
	}
	tests := []struct {
		name    string
		args    args
		want    Authentication
		wantErr bool
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newTelnetConnectionAndAuth(tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTelnetConnectionAndAuth() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTelnetConnectionAndAuth() = %v, want %v", got, tt.want)
			}
		})
	}
}
