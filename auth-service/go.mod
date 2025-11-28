module granula_api/services/auth-service

go 1.23

require (
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/google/uuid v1.5.0
	golang.org/x/crypto v0.18.0
	google.golang.org/grpc v1.60.1
	google.golang.org/protobuf v1.32.0
	gorm.io/driver/postgres v1.5.4
	gorm.io/gorm v1.25.5
	granula_api/shared v0.0.0
)

replace granula_api/shared => ../../shared

