package config

type SSH struct {
	User         string `yaml:"user"`
	Hostname     string `yaml:"hostname"`
	IdentityFile string `yaml:"identity_file"`
	Passphrase   string `yaml:"passphrase"`
}

type Server struct {
	SSH SSH `yaml:"ssh"`
}
