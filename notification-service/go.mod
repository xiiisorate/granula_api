module granula_api/services/notification-service

go 1.23

require (
	github.com/google/uuid v1.5.0
	google.golang.org/grpc v1.60.1
	google.golang.org/protobuf v1.32.0
	gorm.io/driver/postgres v1.5.4
	gorm.io/gorm v1.25.5
	granula_api/shared v0.0.0
)

replace granula_api/shared => ../../shared

