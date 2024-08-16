package rabbitmq

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

const (
	PicklistQueryEvent = "picklistquery.created"
)

func ConnectAmqp(config *Config, l *zap.Logger) (*amqp.Channel, func() error, error) {

	address := fmt.Sprintf("amqp://%s:%s@%s:%s/", config.AmqpUser, config.AmqpPass, config.AmqpHost, config.AmqpPort)

	connection, err := amqp.Dial(address)
	if err != nil {
		l.Error("error initializing", zap.Error(err))
		return nil, nil, err
	}

	channel, err := connection.Channel()
	if err != nil {
		l.Error("error connecting channel", zap.Error(err))
		connection.Close()
		return nil, nil, err
	}

	err = channel.ExchangeDeclare(PicklistQueryEvent, "fanout", true, false, false, false, nil)
	if err != nil {
		l.Error("error declaring picklist query event exchange", zap.Error(err))
		channel.Close()
		connection.Close()
		return nil, nil, err
	}

	return channel, connection.Close, nil
}

func LoadConfig(l *zap.Logger) (*Config, error) {
	if err := godotenv.Load(); err != nil {
		l.Error("error loading .env file")
		return nil, err
	}

	config := &Config{
		AmqpUser:    os.Getenv("amqpUser"),
		AmqpPass:    os.Getenv("amqpPass"),
		AmqpHost:    os.Getenv("amqpHost"),
		AmqpPort:    os.Getenv("amqpPort"),
		JsonDirPath: os.Getenv("jsonDirPath"),
	}

	return config, nil
}
