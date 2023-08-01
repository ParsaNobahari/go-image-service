package main

import (
    "os"
    "log"
    "fmt"
    "sync"
    "imageservice/modules"
    "imageservice/producer"
    amqp "github.com/rabbitmq/amqp091-go"
)

var rabbit_host = os.Getenv("RABBIT_HOST")
var rabbit_port = os.Getenv("RABBIT_PORT")
var rabbit_user = os.Getenv("RABBIT_USER")
var rabbit_password = os.Getenv("RABBIT_PASSWORD")

func failOnError(err error, msg string) {
    if err != nil {
        log.Panicf("%s: %s", msg, err)
    }
}

func main() {

    modules.CreateNewDirectory()

    var wg sync.WaitGroup

    conn, err := amqp.Dial("amqp://" + rabbit_user + ":" + rabbit_password + "@" + rabbit_host + ":" + rabbit_port + "/")
    failOnError(err, "Failed to connect to RabbitMQ")
    defer conn.Close()

    ch, err := conn.Channel()
    failOnError(err, "Failed to open a channel")
    defer ch.Close()

    q, err := ch.QueueDeclare(
        "imageservice", // name
        false,   // durable
        false,   // delete when unused
        false,   // exclusive
        false,   // no-wait
        nil,     // arguments
        )
    failOnError(err, "Failed to declare a queue")

    msgs, err := ch.Consume(
        q.Name, // queue
        "",     // consumer
        true,   // auto-ack
        false,  // exclusive
        false,  // no-local
        false,  // no-wait
        nil,    // args
        )
    failOnError(err, "Failed to register a consumer")

    wg.Add(1)
    go func() {
        defer wg.Done()
        for msg := range msgs {
            url := string(msg.Body)
            if(url[0:7] != "images/") {
                fmt.Println("Received URL:", url)

                wg.Add(1)
                go producer.DownloadAndCompressImage(url, &wg, ch, q)
            }
        }
    }()

    log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
    wg.Wait()
}
