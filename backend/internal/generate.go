package internal

//go:generate mockery

//go:generate swag init -g ../cmd/main.go -o ../docs
//go:generate npm run --prefix ../tools/openapi convert docs/swagger.json -o docs/openapi.json
