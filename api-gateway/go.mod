module github.com/xiiisorate/granula_api/api-gateway

go 1.23

require (
	github.com/gofiber/fiber/v2 v2.52.0
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/google/uuid v1.5.0
	github.com/xiiisorate/granula_api/shared v0.0.0
	google.golang.org/grpc v1.60.1
	google.golang.org/protobuf v1.32.0
)

replace github.com/xiiisorate/granula_api/shared => ../shared
