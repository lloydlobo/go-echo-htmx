package internal

type Config struct {
	ApiUrl         string
	Debug          bool
	DebugSleep     bool
	DebugSleepSecs int
	WithProfiling  bool
}

var ServerConfig = Config{
	ApiUrl:         LookupEnv("API_URL", "https://jsonplaceholder.typicode.com/users"),
	Debug:          true,
	DebugSleep:     false,
	DebugSleepSecs: 2,
	WithProfiling:  false,
}
