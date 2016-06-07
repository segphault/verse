package main

import (
    "fmt"
    "encoding/json"
    "io/ioutil"
    "log"
    "regexp"
    "github.com/methane/rproxy"
	"net/http"
	"os"
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

func (server Server) FindMatchingRule(host string) (Rule, error) {
	for _, rule := range server.Rules {
		if rule.Pattern.MatchString(host) {
			return rule, nil
		}
	}
	
	return Rule{}, fmt.Errorf("Couldn't find a rule to match %s", host)
}

func (rule Rule) Apply(req *http.Request) {
	req.URL.Host = rule.Pattern.ReplaceAllString(req.Host, rule.Binding)
	req.URL.Scheme = "http"
	
	if rule.Scheme != "" {
		req.URL.Scheme = rule.Scheme
	}
}

func (server Server) Run() {
	if server.Static != "" {
		fs := http.FileServer(http.Dir(server.Static))
		http.ListenAndServe(fmt.Sprintf(":%d", server.Port), fs)
	} else {
		director := func(req *http.Request) {
			if rule, err := server.FindMatchingRule(req.Host); err == nil {
				rule.Apply(req)
			} else {
				log.Print(err)
			}
		}
		
		proxy := &rproxy.ReverseProxy{Director: director}
		http.ListenAndServe(fmt.Sprintf(":%d", server.Port), proxy)
	}
}

func main() {
	if len(os.Args[1:]) < 1 {
		log.Fatal("Must specify configuration file")
	}
	
	file, ferr := ioutil.ReadFile(os.Args[1])
	if ferr != nil {
		log.Fatal("Failed to read configuration")
	}
	
	var config Config
	if jerr := json.Unmarshal(file, &config); jerr != nil {
		log.Fatal("Failed to parse JSON:", jerr)
	}

	for _, server := range config {
		go server.Run()
	}
	
	select {}
}