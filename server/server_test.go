package server

import (
	"context"
	"net/http"
	"reflect"
	"testing"
)

func TestNewServer(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			NewServer()
		})
	}
}

func Test_requestDecoder(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		args args
		want People
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := requestDecoder(tt.args.w, tt.args.r); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("requestDecoder() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_run(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			run()
		})
	}
}

func Test_startServer(t *testing.T) {
	type args struct {
		ctx    context.Context
		server *http.Server
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := startServer(tt.args.ctx, tt.args.server); (err != nil) != tt.wantErr {
				t.Errorf("startServer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
