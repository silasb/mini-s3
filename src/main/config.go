package main

type Config struct {
	Server struct {
		Host           string
		Port           string
		RootDomainName string
		Store          string
	}
	RPC struct {
		Host		   string
		Port		   int
	}
}
