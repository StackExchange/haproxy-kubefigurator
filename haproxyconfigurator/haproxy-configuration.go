package haproxyconfigurator

// HaproxyConfig provides an interface to create haproxy configurations
type haproxyConfig struct {
	// Listen IP -> Backend
	listenIPs map[string]map[uint16]*haproxyListener
}

type haproxyListener struct {
	name            string
	mode            string
	sslCertificates []string
	// Hostname -> Backend Target
	hostnameBackends map[string]*HaproxyBackend
	useSSL           bool
}

// HaproxyBackend defines an haproxy backend
type HaproxyBackend struct {
	Name          string
	Backends      []HaproxyBackendTarget
	BalanceMethod string
	UseSSL        bool
	VerifySSL     bool
}

// HaproxyBackendTarget defines a backend target for haproxy
type HaproxyBackendTarget struct {
	Name string
	IP   string
	Port int32
}
