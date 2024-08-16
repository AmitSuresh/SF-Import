module github.com/AmitSuresh/sfdataapp/queue

go 1.22.4

require (
	github.com/AmitSuresh/sfdataapp/rabbitmq v0.0.0-00010101000000-000000000000
	github.com/rabbitmq/amqp091-go v1.10.0
	go.uber.org/zap v1.27.0
)

require (
	github.com/joho/godotenv v1.5.1 // indirect
	go.uber.org/multierr v1.10.0 // indirect
)

replace github.com/AmitSuresh/sfdataapp/rabbitmq => ../rabbitmq
