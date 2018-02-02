package main

import (
	"flag"
	"log"

	"github.com/benlaurie/gds-registers/register"
)

var (
	regname = flag.String("register", "register", "name of register (e.g. 'country')")
)

func main() {
	flag.Parse()

	r, err := register.NewRegister(*regname)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%#v", r)

	i, err := r.Info()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%#v", i)
}
