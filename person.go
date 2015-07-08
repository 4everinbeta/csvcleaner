package main

import "fmt"

type person struct {
	First string `required:"true"`
	Last  string `required:"true"`
	Email string `required:"true"`
	Phone string `required:"true"`
	Bogus string `required:"false"`
}

func (p *person) String() string {
	return fmt.Sprintf("first: %s,\tlast: %s,\temail: %s,\tphone: %s\tbogus: %s\n", p.First, p.Last, p.Email, p.Phone, p.Bogus)
}
