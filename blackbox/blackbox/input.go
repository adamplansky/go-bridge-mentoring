package blackbox

import (
	"net/url"
	"time"
)

func parseURL(rawurl string) url.URL {
	u, err := url.Parse(rawurl)
	if err != nil {
		panic(err)
	}
	return *u
}

func ParseInput() []Job {
	return []Job{
		{
			Name:       "testCase1",
			PeriodTime: time.Second * 3,
			Timeout:    time.Second * 15,
			Targets: []url.URL{
				parseURL("https://prometheus.io/"),
				parseURL("https://www.google.com/"),
				parseURL("https://grafana.com/"),
				parseURL("https://github.com/kubernetes/kubernetes"),
				parseURL("https://kubernetes.io/"),
				parseURL("https://tailscale.com/"),
			},
		},
		{
			Name:       "testCase2",
			PeriodTime: time.Second * 2,
			Timeout:    time.Second * 15,
			Targets: []url.URL{
				parseURL("https://invalid-http-xx.com/"),
				//parseURL("https://twitter.com/home"),
			},
		},
	}
}
