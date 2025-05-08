package servicemanager

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/maximmihin/cb625/internal/balancer/roundrobin"
	weightedroundrobin "github.com/maximmihin/cb625/internal/balancer/roundrobin/weighted"
)

type ServiceManagerConfig struct {
	Services map[string]any `json:"services"` // ключ — имя сервиса, значение - любой поддерживаемый конфиг балансировщика
}

func (s *ServiceManagerConfig) UnmarshalJSON(bytes []byte) error {

	s.Services = make(map[string]any)

	var ttt struct {
		Services map[string]json.RawMessage `json:"services"`
	}

	err := json.Unmarshal(bytes, &ttt)
	if err != nil {
		return err
	}

	for serviceName, message := range ttt.Services {

		var tt struct {
			Algorithm string `json:"algorithm"`
		}
		err := json.Unmarshal(message, &tt)
		if err != nil {
			return err
		}

		switch tt.Algorithm {
		case "round_robin":
			rrConfig := new(roundrobin.Config)
			err = json.Unmarshal(message, rrConfig)
			if err != nil {
				return err
			}
			s.Services[serviceName] = rrConfig
		case "weighted_round_robin":
			wrrConfig := new(weightedroundrobin.Config)
			err = json.Unmarshal(message, wrrConfig)
			if err != nil {
				return err
			}
			s.Services[serviceName] = wrrConfig
		default:
			return errors.New("unsupported algorithm")
		}

	}
	return nil
}

func ParseFromJSON(path string) (*ServiceManagerConfig, error) {
	rawJSON, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := new(ServiceManagerConfig)
	err = json.Unmarshal(rawJSON, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}
