package elasticsearch

type Config struct {
	Addresses []string
	Username  string
	Password  string
	TLS       struct {
		CertFile string
		KeyFile  string
	}
}
