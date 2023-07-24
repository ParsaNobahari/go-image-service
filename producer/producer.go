package producer

import (
    "log"
    "net/http"
    "time"
    "context"
    "strings"
    "sync"
    "github.com/disintegration/imaging"
    amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
    if err != nil {
        log.Panicf("%s: %s", msg, err)
    }
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

func DownloadAndCompressImage(
    url string,
    wg *sync.WaitGroup,
    ch *amqp.Channel,
    q amqp.Queue) {
    defer wg.Done()

    resp, err := http.Get(url)
    if err != nil {
        log.Println(err)
        return
    }
    defer resp.Body.Close()

    decodedImg, err := imaging.Decode(resp.Body)
    if err != nil {
        log.Println(err)
        return
    }

    compressedImg := imaging.Resize(decodedImg, 800, 0, imaging.Lanczos)

    imageAndDirectoryName := "images/" + lastStringAfterSlash(url, "/")

    err = imaging.Save(compressedImg, imageAndDirectoryName)
    if err != nil {
        log.Println(err)
        return
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    err = ch.PublishWithContext(ctx,
        "",     // exchange
        q.Name, // routing key
        false,  // mandatory
        false,  // immediate
        amqp.Publishing{
            ContentType: "text/plain",
            Body:        []byte(imageAndDirectoryName),
        })
    failOnError(err, "Failed to publish a message")
}
