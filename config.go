package main

import (
	"encoding/xml"
	"io/ioutil"
)

type Config struct {
	BindTo string
	SofiaGatewayName string
	SofiaGatewayHost string

	RouteRules []CfgRouteRule `xml:"Routes>Rule"`
}

func LoadConfig(path string) *Config {
	fc, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	cfg := &Config{}
	err = xml.Unmarshal(fc, cfg)
	if err != nil {
		panic(err)
	}

	return cfg
}

func (cfg *Config) RouteRuleForDID(did string) *CfgRouteRule {
	for i := 0; i < len(cfg.RouteRules); i ++ {
		if cfg.RouteRules[i].DID == did {
			return &cfg.RouteRules[i]
		}
	}

	return nil
}

type CfgRouteRule struct {
	DID string `xml:",attr"`
	URL string `xml:",attr"`
}

