package main

import (
    "fmt"
    "encoding/json"
    "io/ioutil"
    "log"
    "github.com/methane/rproxy"
    "net/http"
    "os"
    "rsc.io/letsencrypt"
    "crypto/tls"
)

type Config struct {
	Servers []ServerConfig
	Certs string
}

type Rule struct {
	Pattern Regex
	Binding string
}

type ServerConfig struct {
	Port int
	Rules []Rule
	Static string
	TLS bool
}

func (server ServerConfig) FindMatchingRule(host string) (Rule, error) {
	for _, rule := range server.Rules {
		if rule.Pattern.MatchString(host) {
			return rule, nil
		}
	}
	
	return Rule{}, fmt.Errorf("Couldn't find a rule to match %s", host)
}

func (config ServerConfig) Run(certManager letsencrypt.Manager) {
	director := func(req *http.Request) {
		if rule, err := config.FindMatchingRule(req.Host); err == nil {
			req.URL.Host = rule.Pattern.ReplaceAllString(req.Host, rule.Binding)
			req.URL.Scheme = "http"
		}
	}
	
	server := &http.Server{
		Addr: fmt.Sprintf(":%d", config.Port),
		Handler: &rproxy.ReverseProxy{Director: director},
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
		},
	}
	
	if config.Static != "" {
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