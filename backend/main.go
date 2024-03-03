package main

import (
	"codeflare/api"
	"flag"
	"log"
)

func main() {
	addr := flag.String("addr", ":8080", "server port address")
	flag.Parse()
	svr := api.NewServer(*addr)
	log.Fatal(svr.Run())
}
