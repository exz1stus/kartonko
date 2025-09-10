package main

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"server/internal/env"
	"server/internal/models"

	_ "server/internal/docs"

	_ "ariga.io/atlas-provider-gorm/gormschema"
)

//@title kartonko API
//@version 1.0
//@description kartonko web server
//@securityDefinitions.apikey BearerAuth
//@in cookie
//@name jwt

// swagger:model
type Response gin.H

const (
	frontendOrigin string = "http://localhost:3000"
)

var App *Application

type Application struct {
	Router         *gin.Engine
	Models         *models.Models
	RequestHandler *RequestHandler
	JWTSecret      string
}

func main() {
	models := models.MustInitStorageSqlite()
	reqHandler := MustInitReqHandler(models)

	models.Log.AddEntryType("image_created")

	App := &Application{Models: models, RequestHandler: reqHandler, JWTSecret: env.GetEnvString("JWT_SECRET")}
	App.initRoutes()
	App.Router.Run(fmt.Sprintf("localhost:%d", env.GetEnvInt("PORT")))
}
