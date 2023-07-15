package main

import (
//	"bytes"
	"fmt"
//	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/disintegration/imaging"
	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
    if err != nil {
        log.Panicf("%s: %s", msg, err)
    }
}

func main() {

    conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
    failOnError(err, "Failed to connect to RabbitMQ")
    defer conn.Close()

    ch, err := conn.Channel()
    failOnError(err, "Failed to open a channel")
    defer ch.Close()

    queue, err := ch.QueueDeclare(
        "image_queue", // queue name
        true,          // durable
        false,         // delete when unused
        false,         // exclusive
        false,         // no-wait
        nil,           // arguments
        )
    failOnError(err, "Failed to declare a queue")

    msgs, err := ch.Consume(
        queue.Name, // queue name
        "",         // consumer
        true,       // auto-ack
        false,      // exclusive
        false,      // no-local
        false,      // no-wait
        nil,        // args
        )
    if err != nil {
        log.Fatal(err)
    }

    for msg := range msgs {

        url := string(msg.Body)
        fmt.Println("Received URL:", url)

        resp, err := http.Get(url)
        if err != nil {
            log.Println(err)
            continue
        }
        defer resp.Body.Close()

        img, err := imaging.Decode(resp.Body)
        if err != nil {
            log.Println(err)
            continue
        }

        compressedImg := imaging.Resize(img, 800, 0, imaging.Lanczos)

        err = imaging.Save(compressedImg, "compressed.jpg")
        if err != nil {
            log.Println(err)
            continue
        }
    }
}
