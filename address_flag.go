package main

type addressFlags []string

func (i *addressFlags) String() string {
	return "my string representation"
}

func (i *addressFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}
