package servicemanager

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/maximmihin/cb625/internal/balancer/healthcheck"
	"github.com/maximmihin/cb625/internal/balancer/roundrobin"
	weightedroundrobin "github.com/maximmihin/cb625/internal/balancer/roundrobin/weighted"
)

func TestParseConfig(t *testing.T) {
	type args struct {
		jsonStr []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *ServiceManagerConfig
		wantErr bool
	}{
		{
			name: "green",
			args: args{
				jsonStr: []byte(jsonOne),
			},
			want: &ServiceManagerConfig{
				Services: map[string]any{
					"web_app": &weightedroundrobin.Config{
						BackendUrls: []weightedroundrobin.ServerConfig{
							{URL: "http://web1:8080", Weight: 3},
							{URL: "http://web2:8080", Weight: 2},
						},
					},
					"auth_service": &roundrobin.Config{
						ActiveHealthCheck: &healthcheck.ActiveHealthCheckConfig{
							HealthHandler: "/ping",
							Method:        "GET",
							Interval:      10,
							Timeout:       1,
						},
						BackendUrls: []roundrobin.ServerConfig{
							{URL: "http://auth1:9090"},
							{URL: "http://auth2:9090"},
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := new(ServiceManagerConfig)
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
  "services": {
    "web_app": {
      "algorithm": "weighted_round_robin",
      "servers": [
        { "url": "http://web1:8080", "weight": 3 },
        { "url": "http://web2:8080", "weight": 2 }
      ]
    },
    "auth_service": {
      "algorithm": "round_robin",
      "active_health_check": {
        "health_handler": "/ping",
        "method": "GET",
        "interval": 10,
        "timeout": 1
      },
      "servers": [
        { "url": "http://auth1:9090" },
        { "url": "http://auth2:9090" }
      ]
    }
  }
}`
