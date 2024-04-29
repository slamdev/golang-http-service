package integration

type Config struct {
	Http struct {
		Port int32
	}
	Actuator struct {
		Port int32
	}
	Telemetry struct {
		Logs struct {
			Level  string
			Format string
		}
		Metrics struct {
			Output string // noop, stdout, remote
		}
		Traces struct {
			Output string // noop, stdout, remote
		}
	}
	BaseUrl  string `yaml:"baseUrl"`
	Petstore struct {
		Url string
	}
	Auth struct {
		Enabled              bool
		AllowedIssuers       []string `yaml:"allowedIssuers"`
		JwtSuperuserAudience string   `yaml:"jwtSuperuserAudience"`
		JwkSetUri            string   `yaml:"jwkSetUri"`
		Roles                []struct {
			Name     string
			Audience string
		}
	}
}
