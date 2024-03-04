package main

import "net/url"

type User struct {
	Name  string          `json:"name"`
	Links []HostedProject `json:"links"`
}

type HostedProject struct {
	GithubLink    url.URL `json:"github_link"`
	DirectoryAddr string  `josn:"directory_addr"`
}
