package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/methane/rproxy"
	"rsc.io/letsencrypt"
)

type Config struct {
	Servers []ServerConfig
	Certs   string
}

type Rule struct {
	Pattern string
	Binding string
}

type ServerConfig struct {
	Port   int
	Rules  []Rule
	Static string
	TLS    bool
}

func (server ServerConfig) FindMatchingRule(host string) (string, error) {
	for _, rule := range server.Rules {
		if len(rule.Pattern) > 0 &&
			rule.Pattern[0] == '.' &&
			strings.HasSuffix(host, rule.Pattern) {
			prefix := host[0 : len(host)-len(rule.Pattern)]
			if len(rule.Binding) > 0 && rule.Binding[0] == '.' {
				// e.g., host=www.example.com, pattern=.example.com, binding=.example,
				// then return www.example
				return prefix + rule.Binding, nil
			} else {
				// e.g., host = www.example.com, pattern .example.com, binding=example
				// then return example
				return rule.Binding, nil
			}
		} else if rule.Pattern == host {
			// e.g., host = www.example.com, pattern = www.example.com, binding=example
			// then return example
			return rule.Binding, nil
		}
	}

	return "", fmt.Errorf("Couldn't find a rule to match %s", host)
}

func (config ServerConfig) Run(certManager letsencrypt.Manager) {
	director := func(req *http.Request) {
		if binding, err := config.FindMatchingRule(req.Host); err == nil {
			req.URL.Host = binding
			req.URL.Scheme = "http"
			return
		}
		// TODO intercept request or direct to error responder if we fall
		// through to here.
	}

	server := &http.Server{
		Addr: fmt.Sprintf(":%d", config.Port),
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
		},
	}

	if config.Static == "" {
		server.Handler = &rproxy.ReverseProxy{
			Director:      director,
			FlushInterval: 500,
		}
	} else {
		server.Handler = http.FileServer(http.Dir(config.Static))
	}

	if config.TLS == true {
		server.ListenAndServeTLS("", "")
	} else {
		server.ListenAndServe()
	}
}

func main() {
	if len(os.Args[1:]) < 1 {
		log.Fatal("Must specify configuration file")
	}

	file, ferr := ioutil.ReadFile(os.Args[1])
	if ferr != nil {
		log.Fatal("Failed to read configuration:", ferr)
	}

	var config Config
	if jerr := json.Unmarshal(file, &config); jerr != nil {
		log.Fatal("Failed to parse JSON:", jerr)
	}

	var certManager letsencrypt.Manager
	if config.Certs != "" {
		if err := certManager.CacheFile(config.Certs); err != nil {
			log.Fatal(err)
		}
	}

	for _, server := range config.Servers {
		go server.Run(certManager)
	}

	select {}
}
