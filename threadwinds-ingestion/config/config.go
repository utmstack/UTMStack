package config

type TWConfig struct {
	InternalKey         string
	BackendURL          string
	CustomersManagerURL string
	ThreadWindsURL      string
	OpenSearchHost      string
	OpenSearchPort      string
	DBHost              string
	DBPort              string
	DBUser              string
	DBPassword          string
	DBName              string
}

func GetTWConfig() (*TWConfig, error) {
	cfg := &TWConfig{
		InternalKey:         GetInternalKey(),
		BackendURL:          GetBackendUrl(),
		CustomersManagerURL: GetCustomersManagerUrl(),
		ThreadWindsURL:      GetThreadWindsURL(),
		OpenSearchHost:      GetOpenSearchHost(),
		OpenSearchPort:      GetOpenSearchPort(),
		DBHost:              GetDBHost(),
		DBPort:              GetDBPort(),
		DBUser:              GetDBUser(),
		DBPassword:          GetDBPassword(),
		DBName:              GetDBName(),
	}

	return cfg, nil
}
