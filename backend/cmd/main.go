package main

import (
	_ "server/docs"
	"server/internal/api"

	_ "ariga.io/atlas-provider-gorm/gormschema"
)

//@title kartonko API
//@version 1.0
//@description kartonko web server
//@securityDefinitions.apikey BearerAuth
//@in cookie
//@name jwt

func main() {
	// media.RegenerateThumbnails(env.GetEnvString("UPLOADS_PATH"), env.GetEnvString("THUMBNAILS_PATH"))
	api := api.MustInitApi()
	api.Run()
}
