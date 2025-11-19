package main

import "github.com/gsaaraujo/pay-bank-api/internal"

func main() {
	internal.NewHttpServer().Start()
}
