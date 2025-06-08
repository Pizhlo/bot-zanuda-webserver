package config

import (
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v2"
)

type Server struct {
	Address string `yaml:"address" validate:"required,hostname_port"`
}

type RabbitMQ struct {
	Address    string `yaml:"address" validate:"required,rabbitmq_address"`
	NoteQueue  string `yaml:"note_queue" validate:"required"`
	SpaceQueue string `yaml:"space_queue" validate:"required"`
}

type Postgres struct {
	Host     string `yaml:"host" validate:"required,hostname"`
	Port     int    `yaml:"port" validate:"required,min=1024,max=65535"`
	User     string `yaml:"user" validate:"required"`
	Password string `yaml:"password" validate:"required"`
	DBName   string `yaml:"db_name" validate:"required"`
}

type ElasticSearch struct {
	Address string `yaml:"address" validate:"required,url"`
}

type Redis struct {
	Address string `yaml:"address" validate:"required,hostname_port"`
}

type Auth struct {
	SecretKey string `yaml:"secret_key" validate:"required,min=32"`
}

type Config struct {
	Server Server `yaml:"server"`

	LogLevel string `yaml:"log_level" validate:"required,oneof=debug info warn error"`

	Storage struct {
		Postgres      Postgres      `yaml:"postgres"`
		ElasticSearch ElasticSearch `yaml:"elasticsearch"`
		Redis         Redis         `yaml:"redis"`
		RabbitMQ      RabbitMQ      `yaml:"rabbitmq"`
	} `yaml:"storage"`

	Auth Auth `yaml:"auth"`
}

func LoadConfig(path string) (*Config, error) {
	cfg := &Config{}

	// Читаем YAML файл
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Парсим YAML
	if err := yaml.Unmarshal(yamlFile, cfg); err != nil {
		return nil, err
	}

	// Создаем валидатор
	validate := validator.New()

	validate.RegisterValidation("rabbitmq_address", ValidateRabbitMQAddress)

	// Валидируем конфиг
	if err := validate.Struct(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// ValidateRabbitMQAddress implements validator.Func
func ValidateRabbitMQAddress(fl validator.FieldLevel) bool {
	return strings.HasPrefix(fl.Field().String(), "amqp://")
}
