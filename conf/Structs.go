package conf

type RemoteServer struct {
	Host           string
	Port           int
	IsPrivateKey   bool
	Username       string
	Passwd         string
	PrivateKeyFile string
	Path           string
}

type Conf struct {
	SkipExist bool
	LocalPath string
	Exclude   []string
	Remote    []RemoteServer
}
