package main

import (
	"context"
	"fmt"
	"os"
	"time"

	probe "github.com/probe/lib"
)

func main() {
	//TODO: read duration, rate, target from command line args
	duration := time.Duration(5) * time.Second
	rate := 10
	target := "https://gitlab.com"

	p, err := probe.New(duration, rate, target)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ctx := context.TODO()
	fmt.Println(fmt.Sprintf("Measuring response times for = %s, execution time = %v ...", target, duration))
	res := p.Start(ctx)
	fmt.Println(fmt.Sprintf("Statistics for = %s:", target))
	err = res.Print(os.Stdout)
	if err != nil {
		fmt.Printf("sorry something went wrong while printing the results, details = %s\n", err.Error())
		os.Exit(1)
	}
}
