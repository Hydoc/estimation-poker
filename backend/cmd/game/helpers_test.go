package main

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"

	"github.com/Hydoc/estimation-poker/backend/internal/assert"
)

func Test_readIdParam(t *testing.T) {
	tests := []struct {
		name   string
		params httprouter.Params
		want   uuid.UUID
		err    error
	}{
		{
			name: "should return correct id param",
			params: httprouter.Params{
				{
					Key:   "roomId",
					Value: "2671ca6e-9573-4294-9972-26bed5c92629",
				},
			},
			want: uuid.MustParse("2671ca6e-9573-4294-9972-26bed5c92629"),
			err:  nil,
		},
		{
			name: "should return nil and error for invalid room id",
			params: httprouter.Params{
				{
					Key:   "roomId",
					Value: "invalid",
				},
			},
			want: uuid.Nil,
			err:  errors.New("invalid roomId param"),
		},
		{
			name:   "should return nil and error for missing room id",
			params: httprouter.Params{},
			want:   uuid.Nil,
			err:    errors.New("invalid roomId param"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/does/not/matter", nil)
			req = req.WithContext(context.WithValue(req.Context(), httprouter.ParamsKey, tt.params))
			got, err := readIdParam(req)

			assert.DeepEqual(t, got.String(), tt.want.String())
			assert.DeepEqual(t, err, tt.err)
		})
	}
}

func Test_readNameQueryParm(t *testing.T) {
	tests := []struct {
		name string
		want string
		url  string
		err  error
	}{
		{
			name: "should return correct name",
			url:  "/v1/hello?name=Tester",
			want: "Tester",
			err:  nil,
		},
		{
			name: "should return an empty string and an error for invalid name",
			url:  "/v1/hello?name=%20",
			want: "",
			err:  errors.New("invalid name query parameter"),
		},
		{
			name: "should return nil and error for missing name",
			url:  "/v1/hello",
			want: "",
			err:  errors.New("invalid name query parameter"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			got, err := readNameQueryParam(req)

			assert.DeepEqual(t, got, tt.want)
			assert.DeepEqual(t, err, tt.err)
		})
	}
}
