package main

import (
	_ "server/docs"
	"server/internal/api"
)

//@title kartonko API
//@version 1.0
//@description kartonko web server
//@securityDefinitions.apikey BearerAuth
//@in cookie
//@name jwt

func main() {
	api := api.MustInitApi()
	api.Run()
}
