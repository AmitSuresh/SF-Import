module github.com/AmitSuresh/sfdataapp

go 1.22.4

require (
	github.com/AmitSuresh/sfdataapp/rabbitmq v0.0.0-00010101000000-000000000000
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/gorilla/mux v1.8.1
	github.com/joho/godotenv v1.5.1
	github.com/rabbitmq/amqp091-go v1.10.0
	go.uber.org/zap v1.27.0
)

require go.uber.org/multierr v1.10.0 // indirect

replace github.com/AmitSuresh/sfdataapp/rabbitmq => /rabbitmq
