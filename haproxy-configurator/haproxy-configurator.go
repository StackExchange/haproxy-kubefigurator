package haproxyConfigurator

import (
	"sort"
	"strconv"

	"github.com/fatih/color"
)

// HaproxyConfigurator provides an interface to dynamically generate haproxy configs
type HaproxyConfigurator struct {
	desiredConfig haproxyConfig
}

// Initialize sets up a new HaproxyConfigurator
func (h *HaproxyConfigurator) Initialize() {
	h.desiredConfig.listenIPs = make(map[string]map[uint16]*haproxyListener)
}

// HaproxyListener structure provides configuration options
type HaproxyListenerConfig struct {
	Name             string
	Backend          HaproxyBackend
	Hostname         string
	ListenIP         string
	ListenPort       uint16
	Mode             string
	SslCertificate   string
	validationErrors []string
}

func (hlc *HaproxyListenerConfig) addValidationError(message string) {
	hlc.validationErrors = append(hlc.validationErrors, message)
}

func (hlc *HaproxyListenerConfig) validate(h *HaproxyConfigurator) bool {
	// Default to validated
	var validated = true

	// Check mode
	if hlc.Mode != "http" && hlc.Mode != "tcp" {
		hlc.addValidationError("Invalid mode (" + hlc.Mode + ") specified - valid options 'http', 'tcp'")
		validated = false
	}

	// Against other services' configurations
	if _, exists := h.desiredConfig.listenIPs[hlc.ListenIP]; exists {
		if _, exists := h.desiredConfig.listenIPs[hlc.ListenIP][hlc.ListenPort]; exists {
			// Validate Mode
			if h.desiredConfig.listenIPs[hlc.ListenIP][hlc.ListenPort].mode != hlc.Mode {
				hlc.addValidationError("Mode does not match on service")
				validated = false
			}

			// Validate SSL
			if h.desiredConfig.listenIPs[hlc.ListenIP][hlc.ListenPort].useSSL != (hlc.SslCertificate != "") {
				hlc.addValidationError("SSL Certificate provided on a service that isn't using SSL")
				validated = false
			}
		}
	}

	return validated
}

// AddListener to haproxy
func (h *HaproxyConfigurator) AddListener(
	hlc HaproxyListenerConfig,
) {
	if hlc.validate(h) {
		if _, exists := h.desiredConfig.listenIPs[hlc.ListenIP]; !exists {
			h.desiredConfig.listenIPs[hlc.ListenIP] = make(map[uint16]*haproxyListener)
		}

		if _, exists := h.desiredConfig.listenIPs[hlc.ListenIP][hlc.ListenPort]; !exists {
			var listener = haproxyListener{
				name:             hlc.Name,
				mode:             hlc.Mode,
				sslCertificates:  []string{},
				hostnameBackends: make(map[string]*HaproxyBackend),
				useSSL:           hlc.SslCertificate != "",
			}
			h.desiredConfig.listenIPs[hlc.ListenIP][hlc.ListenPort] = &listener
		} else {
			// Don't allow duplicate listeners on TCP endpoints
			if hlc.Mode == "tcp" {
				color.Red("A listener for another TCP service is already configured on the port (" + strconv.Itoa(int(hlc.ListenPort)) + ")")
				panic("This could cause unpredictability in the service router")
			}
		}

		if hlc.SslCertificate != "" {
			h.desiredConfig.listenIPs[hlc.ListenIP][hlc.ListenPort].sslCertificates = append(h.desiredConfig.listenIPs[hlc.ListenIP][hlc.ListenPort].sslCertificates, hlc.SslCertificate)
		}

		if hlc.Mode == "tcp" {
			hlc.Hostname = "_"
		}
		h.desiredConfig.listenIPs[hlc.ListenIP][hlc.ListenPort].hostnameBackends[hlc.Hostname] = &hlc.Backend
	} else {
		color.Red(hlc.Name)
		for _, message := range hlc.validationErrors {
			color.Red("  " + message)
		}
	}
}

// Render the haproxy configuration
func (h *HaproxyConfigurator) Render() string {
	var config = ""

	// Build Front-Ends
	for listenIP := range h.desiredConfig.listenIPs {
		for port, listener := range h.desiredConfig.listenIPs[listenIP] {
			config += "frontend " + listener.name + "\n"
			config += "    mode " + listener.mode + "\n"
			config += "    bind " + listenIP + ":" + strconv.Itoa(int(port))
			if listener.useSSL {
				sort.Strings(listener.sslCertificates)
				config += " ssl"
				var previous = ""
				for _, certificate := range listener.sslCertificates {
					if previous != certificate {
						config += " crt " + certificate
						previous = certificate
					}
				}
				config += "\n"
				if listener.mode == "http" {
					config += "    reqadd x-forwarded-proto:\\ https"
				}
			}
			config += "\n"
			config += "\n"
			if listener.mode == "http" {
				for hostname, backend := range listener.hostnameBackends {
					config += "    # Set up backend selection for " + hostname + "\n"
					config += "    use_backend " + backend.Name + " if { hdr(host) -i " + hostname + " }\n"
					config += "    use_backend " + backend.Name + " if { hdr(host) -i " + hostname + ":" + strconv.Itoa(int(port)) + " }\n"
				}
			} else if listener.mode == "tcp" {
				config += "    # Set up default_backend\n"
				config += "    default_backend " + listener.hostnameBackends["_"].Name + "\n"
			}
			config += "\n"
		}
	}

	// Build Back-Ends
	for listenIP := range h.desiredConfig.listenIPs {
		for _, listener := range h.desiredConfig.listenIPs[listenIP] {
			for _, backend := range listener.hostnameBackends {
				config += "backend " + backend.Name + "\n"
				config += "    mode " + listener.mode + "\n"
				config += "    balance " + backend.BalanceMethod + "\n"
				config += "\n"
				config += "    # Backend Servers\n"
				for _, backendServer := range backend.Backends {
					config += "    server " + backendServer.Name + " " + backendServer.IP + ":" + strconv.Itoa(int(backendServer.Port))
					config += " check"
					if backend.UseSSL {
						config += " ssl"
						if !backend.VerifySSL {
							config += " verify none"
						}
					}
					config += "\n"
				}
				config += "\n"
			}
		}
	}

	return config
}
