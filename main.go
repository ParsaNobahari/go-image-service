package main

import (
    "fmt"
    "log"
    "net/http"
    "os"
    "github.com/google/uuid"
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

func lastStringAfterSlash(value string, a string) string {
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

    var uuids []string

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

        out:
        for msg := range msgs {

            senderUUID := uuid.New().String()

            fmt.Println("created new UUID")
            if uuids != nil {
                if uuids[len(uuids)-1] == msg.Headers["uuid"] {
                    fmt.Println("duplicated detected, ignoring")
                    continue out
                } else {
                    uuids = append(uuids, senderUUID)
                    fmt.Println("new UUID has been added")
                }
            } else {
                uuids = append(uuids, senderUUID)
                fmt.Println("first UUID added")
            }

            url := string(msg.Body)
            fmt.Println("Received URL:", url)

            resp, err := http.Get(url)
            if err != nil {
                log.Println(err)
                continue
            }
            defer resp.Body.Close()

            decodedImg, err := imaging.Decode(resp.Body)
            if err != nil {
                log.Println(err)
                continue
            }

            compressedImg := imaging.Resize(decodedImg, 800, 0, imaging.Lanczos)

            imageName, err := getPageName(url)
            if err != nil {
                log.Println(err)
            }

            imageAndDirectoryName :=  "images/" + lastStringAfterSlash(string(imageName), "/")

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
                    Headers:     amqp.Table{
                        "uuid": uuids[len(uuids)-1],
                    },
                })
            failOnError(err, "Failed to publish a message")
        }
    }()

    log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
    <-forever
}
