package main

import (
	"context"
	"fmt"
	"os"
	"time"

	probe "github.com/probe/lib"
)

func main() {
	fmt.Println("Hello my name is probe. I was designed to measure your server response times")

	//TODO: read duration, rate, target from command line args
	duration := time.Duration(30) * time.Second
	rate := 2
	target := ""
	ctx := context.TODO()

	//TODO: make sure that duration, rate, target are reasonable.
	p := probe.New()

	res := p.Start(ctx, duration, rate, target)
	err := res.Print(os.Stdout)
	if err != nil {
		fmt.Printf("sorry something went wrong while printing the results, details = %s\n", err.Error())
		os.Exit(1)
	}
}
