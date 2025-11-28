module granula_api/services/api-gateway

go 1.23

require (
	github.com/gofiber/fiber/v2 v2.52.0
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/google/uuid v1.5.0
	google.golang.org/grpc v1.60.1
	granula_api/shared v0.0.0
)

replace granula_api/shared => ../../shared
