package data

type Config struct {
	Address    string   `yaml:"address"`
	Network    string   `yaml:"network"`
	Netmask    int      `yaml:"netmask"`
	ListenPort int      `yaml:"listen_port"`
	Users      []string `yaml:"users"`
}
