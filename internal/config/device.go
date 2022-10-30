package config

type Device struct {
	Address    string `yaml:"address"`
	Network    string `yaml:"network"`
	Netmask    int    `yaml:"netmask"`
	DNS        string `yaml:"dns"`
	ListenPort int    `yaml:"listen_port"`

	PostUp  string `yaml:"post_up"`
	PreDown string `yaml:"pre_down"`

	Users []User `yaml:"users"`
}

type User struct {
	Name       string   `yaml:"name"`
	Address    string   `yaml:"address"`
	AllowedIPs []string `yaml:"allowed_ips"`
}
