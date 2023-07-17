package main

import (
    //	"bytes"
    "fmt"
    //	"io"
    "log"
    "net/http"
    "os"
    "strings"
    "net/url"
    "time"
    "context"
    "github.com/disintegration/imaging"
    amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
    if err != nil {
        log.Panicf("%s: %s", msg, err)
    }
}

func getPageName(URL string) (string, error) {
    u, err := url.Parse(URL)
    if err != nil {
        return "", err
    }
    return u.Path[1:], nil
}

func after(value string, a string) string {
    pos := strings.LastIndex(value, a)
    if pos == -1 {
        return ""
    }
    adjustedPos := pos + len(a)
    if adjustedPos >= len(value) {
        return ""
    }
    return value[adjustedPos:]
}

func main() {

    if _, err := os.Stat("images"); err != nil {
        if os.IsNotExist(err) {
            if err := os.Mkdir("images", os.ModePerm);
            err != nil {
                log.Fatal(err)
            }
        }
    }

    conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
    failOnError(err, "Failed to connect to RabbitMQ")
    defer conn.Close()

    ch, err := conn.Channel()
    failOnError(err, "Failed to open a channel")
    defer ch.Close()

    q, err := ch.QueueDeclare(
        "hello", // name
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

    var forever chan struct{}

    go func() {
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

            name, err := getPageName(url)
            if err != nil {
                log.Println(err)
            }

            imageAndDirectoryName :=  "images/" + after(string(name), "/")

            err = imaging.Save(compressedImg, imageAndDirectoryName)
            if err != nil {
                log.Println(err)
                continue
            }

            ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
            defer cancel()

            err = ch.PublishWithContext(ctx,
                "",     // exchange
                q.Name, // routing key
                false,  // mandatory
                false,  // immediate
                amqp.Publishing {
                    ContentType: "text/plain",
                    Body:        []byte(imageAndDirectoryName),
                })
            failOnError(err, "Failed to publish a message")
        }
    }()

    log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
    <-forever
}
