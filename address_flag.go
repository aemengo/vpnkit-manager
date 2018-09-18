package main

import "strings"

type addressFlags []string

func (i *addressFlags) String() string {
	return strings.Join(*i, " ")
}

func (i *addressFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}
