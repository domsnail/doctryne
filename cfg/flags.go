package cfg

import "flag"

func NewConfigFromFlags() (config *Config, err error) {
	config = new(Config)

	// === Configuration file
	//configFile

	// === Server setup
	server := flag.Bool("server", false, "use server mode")
	host := flag.String("host", "0.0.0.0", "server address")
	port := flag.Uint64("port", 80, "server port")
	accessKey := flag.String("access-key", "", "server access key")

	flag.Parse()

	if server != nil && *server {
		config.Server.Enabled = true

		config.Server.Host = *host
		config.Server.Port = uint32(*port)

		config.Server.AccessKey = *accessKey
	}

	return config, nil
}
