package main

import (
    "fmt"
    "encoding/json"
    "io/ioutil"
    "log"
    "regexp"
    "github.com/methane/rproxy"
	"net/http"
)

type Regex struct {
	*regexp.Regexp
}

func (r *Regex) UnmarshalJSON(b []byte) error {
	str := new(string)
	json.Unmarshal(b, str)
	
	compiled, err := regexp.Compile(*str)
	
	if err != nil {
		return err
	}
	
	r.Regexp = compiled
	return nil
}

type Config []Server

type Rule struct {
	Pattern Regex
	Binding string
	Scheme string
}

type Server struct {
	Port int
	Rules []Rule
	Static string
}

func (server Server) Run() {
	if server.Static != "" {
		fs := http.FileServer(http.Dir(server.Static))
		http.ListenAndServe(fmt.Sprintf(":%d", server.Port), fs)
	} else {
		director := func(req *http.Request) {
			for _, rule := range server.Rules {
				if rule.Pattern.MatchString(req.Host) {
					req.URL.Host = rule.Pattern.ReplaceAllString(req.Host, rule.Binding)
					fmt.Println(fmt.Sprintf("%s -> %s", req.Host, req.URL.Host))
					
					if rule.Scheme != "" {
						req.URL.Scheme = rule.Scheme
					} else {
						req.URL.Scheme = "http"
					}
				}
			}
		}
		
		proxy := &rproxy.ReverseProxy{Director: director}
		http.ListenAndServe(fmt.Sprintf(":%d", server.Port), proxy)
	}
}

func main() {
	file, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatal("Failed to read configuration")
	}
	
	var config Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		log.Fatal("Failed to parse JSON:", err)
	}

	for _, server := range config {
		go server.Run()
	}
	
	select {}
}