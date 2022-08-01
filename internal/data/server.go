package data

type Server struct {
	Host       string `json:"host"`
	Name       string `json:"name"`
	ListenPort int    `json:"listen_port"`

	Address string `json:"address"`
	Network string `json:"network"`
	Netmask int    `json:"netmask"`

	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`

	LastAddress string `json:"last_address"`
}
