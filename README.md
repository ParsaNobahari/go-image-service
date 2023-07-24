## Go Image Service with RabbitMQ

### Usage

#### Run RabbitMQ Server

    docker run -d --hostname rabbitmq-server --name image-service -p 15672:15672 -p 5672:5672 rabbitmq:3-management .

#### Build Docker Image

    docker build -t image-service .

#### Run Image Service

    docker run -p 5672:5672 image-service

#### Run Test

    go test
