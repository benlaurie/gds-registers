package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/benlaurie/gds-registers/register"
)

type lister struct {
}

func (*lister) Process(e map[string]interface{}) error {
	t := e["key"].(string)
	fmt.Printf("Register: %s\n", t)
	r, err := register.NewRegister(t)
	if err != nil {
		return err
	}
	//fmt.Printf("%#v\n", r)
	i, err := r.Info()
	if err != nil {
		return err
	}
	fmt.Printf("  Description: %s\n", i.Text)
	fmt.Printf("  Records: %d\n  Entries: %d\n", i.Records, i.Entries)
	fmt.Printf("  Last updated: %s\n", i.LastUpdated.Format("Mon Jan 2 15:04:05 -0700 MST 2006"))
	return nil
}

func main() {
	flag.Parse()

	r, err := register.NewRegister("register")
	if err != nil {
		log.Fatal(err)
	}
	err = r.GetSummaryEntries(&lister{})
	if err != nil {
		log.Fatal(err)
	}
}
