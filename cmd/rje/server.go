package main

import (
	"fmt"
	"os"

	"github.com/orsinium-labs/cliff"
)

func serve(args []string) error {
	flags, err := cliff.Parse(os.Stderr, args, serveFlags)
	if err != nil {
		return err
	}

	fmt.Printf("serve not implemented: %d", flags.port)
	return nil
}
