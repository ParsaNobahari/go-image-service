# Go Image Service with RabbitMQ

## Usage

### Run Local Docker Network

    docker network create rabbits

### Run RabbitMQ in Docker

    docker run -d --rm --net rabbits -p 8080:15672 --hostname rabbit-1 --name image-service rabbitmq:3.8

**note**: if you cannot access rabbitmq management, try this:

#### Enable rabbitmq management

    exec -it rabbit-1 bash
    rabbitmq-plugins enable rabbitmq_maangement

to check if this worked you can simply try:
> rabbitmq-plugins list 

if _rabbitmq_web_dispatch_, _rabbitmq_management_ and _rabbitmq_management_agent_ didn't get enabled, DIY.

### Build Docker Image

    docker build -t image-service .

### Run Image Service

    docker run -p 5672:5672 image-service

### To Run Test Locally

    go test
