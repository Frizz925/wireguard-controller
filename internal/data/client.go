package data

type Client struct {
	PrivateKey   string `json:"private_key"`
	PublicKey    string `json:"public_key"`
	PresharedKey string `json:"preshared_key"`
}
