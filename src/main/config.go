package main

type Config struct {
	Server struct {
		Host           string
		Port           string
		RootDomainName string
		Store          string
	}
}
