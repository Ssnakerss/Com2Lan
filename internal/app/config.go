package app

import "flag"

type Config struct {
	PortName    string
	ClientId    string
	ChannelName string
	RabbitURL   string
}

func (acfg *Config) Read() {
	flag.StringVar(&acfg.PortName, "port-name", "", "port name")
	flag.StringVar(&acfg.ClientId, "client-id", "", "client id")
	flag.StringVar(&acfg.ChannelName, "channel-name", "", "channel name")
	flag.StringVar(&acfg.RabbitURL, "rabbit-url", "amqp://guest:guest@localhost:5672/", "rabbit url")
	flag.Parse()
}
