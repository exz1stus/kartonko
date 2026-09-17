package main

import (
	_ "server/docs"
	"server/internal/api"
)

//@title kartonko API
//@version 1.0
//@description kartonko web server
//@schemes         http https
//
//@securityDefinitions.apikey BearerAuth
//@in cookie
//@name jwt
//@description Bearer JWT authentication stored in the jwt cookie.

func main() {
	api := api.MustInitApi()
	api.Run()
}
