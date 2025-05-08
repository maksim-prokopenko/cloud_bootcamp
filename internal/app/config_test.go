package app

import (
	"encoding/json"
	"github.com/maximmihin/cb625/internal/limiter/tokenbucket"
	"reflect"
	"testing"

	"github.com/maximmihin/cb625/internal/balancer/roundrobin"
	"github.com/maximmihin/cb625/internal/healthcheck"
)

func TestParseConfig(t *testing.T) {
	type args struct {
		jsonStr []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *Config
		wantErr bool
	}{
		{
			name: "green",
			args: args{
				jsonStr: []byte(jsonOne),
			},
			want: &Config{
				Balancer: &roundrobin.Config{
					BackendUrls: []roundrobin.ServerConfig{
						{URL: "http://localhost:8081", Weight: 3}, // TODO check all vaules (weight actually)
						{URL: "http://localhost:8082", Weight: 2},
					},
				},
				Limiter: &tokenbucket.Config{
					ClientsLimits: map[string]float64{
						"vasya": 1.5,
						"petya": 10,
					},
				},
				ActiveHealthCheck: &healthcheck.ActiveHealthCheckConfig{
					HealthHandler: "/ping",
					Method:        "GET",
					Interval:      10,
					Timeout:       100,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := new(Config)
			err := json.Unmarshal(tt.args.jsonStr, got)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal error = %v, wantErr %v", err, tt.wantErr) // TODO ???
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}

const jsonOne = `{
  "balancer": {
    "type": "round_robin",
    "servers": [
      { "url": "http://localhost:8081", "weight": 3 },
      { "url": "http://localhost:8082", "weight": 2 }
    ]
  },
  "limiter": {
    "type": "token_buket",
    "users": [
      { "token": "vasya", "limit":  1.5 },
      { "token": "petya", "limit":  10 }
    ]
  },
  "active_health_check": {
    "handler": "/ping",
    "method": "GET",
    "interval": 10,
    "timeout": 100
  }
}`
