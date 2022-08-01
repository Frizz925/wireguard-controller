package data

type Client struct {
	Name    string `json:"name"`
	Comment string `json:"comment"`
	Address string `json:"address"`

	PrivateKey   string `json:"private_key"`
	PublicKey    string `json:"public_key"`
	PresharedKey string `json:"preshared_key"`
}
