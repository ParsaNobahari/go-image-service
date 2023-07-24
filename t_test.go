package main

import (
    "context"
    "bytes"
    "log"
    "time"
    "github.com/disintegration/imaging"
    "testing"
    "imageservice/modules"
    "os"
    "io/ioutil"
    amqp "github.com/rabbitmq/amqp091-go"
)
func TestImageService(t *testing.T) {
    // Create a temporary directory for storing images
    tempDir, err := ioutil.TempDir("", "images")
    if err != nil {
        t.Fatalf("Failed to create temporary directory: %v", err)
    }
    defer os.RemoveAll(tempDir)

    // Connect to RabbitMQ
    conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
    if err != nil {
        t.Fatalf("Failed to connect to RabbitMQ: %v", err)
    }
    defer conn.Close()

    ch, err := conn.Channel()
    if err != nil {
        t.Fatalf("Failed to open a channel: %v", err)
    }
    defer ch.Close()

    q, err := ch.QueueDeclare(
        "imageservice", // name
        false,   // durable
        false,   // delete when unused
        false,   // exclusive
        false,   // no-wait
        nil,     // arguments
    )
    if err != nil {
        t.Fatalf("Failed to declare a queue: %v", err)
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    body := []string{
        "https://www.kasandbox.org/programming-images/avatars/leaf-blue.png",
        "https://www.kasandbox.org/programming-images/avatars/leaf-green.png",
        "https://www.kasandbox.org/programming-images/avatars/leaf-grey.png",
        "https://www.kasandbox.org/programming-images/avatars/leaf-orange.png",
        "https://www.kasandbox.org/programming-images/avatars/leaf-red.png",
        "https://www.kasandbox.org/programming-images/avatars/leaf-yellow.png",
        "https://www.kasandbox.org/programming-images/avatars/cs-hopper-happy.png",
        "https://www.kasandbox.org/programming-images/avatars/cs-hopper-cool.png",
        "https://www.kasandbox.org/programming-images/avatars/leafers-seed.png",
        "https://www.kasandbox.org/programming-images/avatars/leafers-seedling.png",
        "https://www.kasandbox.org/programming-images/avatars/leafers-sapling.png",
        "https://www.kasandbox.org/programming-images/avatars/leafers-tree.png",
        "https://www.kasandbox.org/programming-images/avatars/leafers-ultimate.png",
        "https://www.kasandbox.org/programming-images/avatars/piceratops-seed.png",
        "https://www.kasandbox.org/programming-images/avatars/piceratops-seedling.png",
        "https://www.kasandbox.org/programming-images/avatars/piceratops-sapling.png",
        "https://www.kasandbox.org/programming-images/avatars/piceratops-tree.png",
        "https://www.kasandbox.org/programming-images/avatars/piceratops-ultimate.png",
        "https://www.kasandbox.org/programming-images/avatars/aqualine-seed.png",
        "https://www.kasandbox.org/programming-images/avatars/aqualine-seedling.png",
        "https://www.kasandbox.org/programming-images/avatars/aqualine-sapling.png",
        "https://www.kasandbox.org/programming-images/avatars/aqualine-tree.png",
        "https://www.kasandbox.org/programming-images/avatars/aqualine-ultimate.png",
        "https://www.kasandbox.org/programming-images/avatars/starky-seed.png",
        "https://www.kasandbox.org/programming-images/avatars/starky-seedling.png",
        "https://www.kasandbox.org/programming-images/avatars/starky-sapling.png",
        "https://www.kasandbox.org/programming-images/avatars/starky-tree.png",
        "https://www.kasandbox.org/programming-images/avatars/starky-ultimate.png",
        "https://www.kasandbox.org/programming-images/avatars/primosaur-seed.png",
        "https://www.kasandbox.org/programming-images/avatars/primosaur-seedling.png",
        "https://www.kasandbox.org/programming-images/avatars/primosaur-sapling.png",
        "https://www.kasandbox.org/programming-images/avatars/primosaur-tree.png",
        "https://www.kasandbox.org/programming-images/avatars/primosaur-ultimate.png",
        "https://www.kasandbox.org/programming-images/avatars/duskpin-seed.png",
        "https://www.kasandbox.org/programming-images/avatars/duskpin-seedling.png",
        "https://www.kasandbox.org/programming-images/avatars/duskpin-sapling.png",
        "https://www.kasandbox.org/programming-images/avatars/duskpin-tree.png",
        "https://www.kasandbox.org/programming-images/avatars/duskpin-ultimate.png",
        "https://www.kasandbox.org/programming-images/avatars/old-spice-man.png",
        "https://www.kasandbox.org/programming-images/avatars/old-spice-man-blue.png",
        "https://www.kasandbox.org/programming-images/avatars/orange-juice-squid.png",
        "https://www.kasandbox.org/programming-images/avatars/purple-pi.png",
        "https://www.kasandbox.org/programming-images/avatars/purple-pi-teal.png",
        "https://www.kasandbox.org/programming-images/avatars/purple-pi-pink.png",
        "https://www.kasandbox.org/programming-images/avatars/spunky-sam.png",
        "https://www.kasandbox.org/programming-images/avatars/spunky-sam-green.png",
        "https://www.kasandbox.org/programming-images/avatars/mr-pants.png",
        "https://www.kasandbox.org/programming-images/avatars/mr-pants-green.png",
        "https://www.kasandbox.org/programming-images/avatars/mr-pants-purple.png",
        "https://www.kasandbox.org/programming-images/avatars/marcimus.png",
        "https://www.kasandbox.org/programming-images/avatars/marcimus-red.png",
        "https://www.kasandbox.org/programming-images/avatars/marcimus-orange.png",
        "https://www.kasandbox.org/programming-images/avatars/marcimus-purple.png",
    }

    for _, b := range body {
        err = ch.PublishWithContext(ctx,
            "",     // exchange
            q.Name, // routing key
            false,  // mandatory
            false,  // immediate
            amqp.Publishing {
                ContentType: "text/plain",
                Body:        []byte(b),
            })
        failOnError(err, "Failed to publish a message")
        log.Printf(" [x] Sent %s\n", b)
    }

    var imageFile string
    msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
    if err != nil {
        t.Fatalf("Failed to register a consumer: %v", err)
    }

    out:
    for msg := range msgs {
        for _, b := range body {
            if bytes.Equal(msg.Body, []byte("images/" + modules.LastStringAfterSlash(b, "/"))) {
                imageFile = string(msg.Body)
                break out
            }
        }
    }

    // Verify that the image was downloaded and saved
    if _, err := os.Stat(imageFile); err != nil {
        t.Fatalf("Failed to find saved image: %v", err)
    }

    // Verify that the image was compressed
    img, err := imaging.Open(imageFile)
    if err != nil {
        t.Fatalf("Failed to open saved image: %v", err)
    }
    if img.Bounds().Dx() > 800 || img.Bounds().Dy() > 800 {
        t.Fatalf("Image was not compressed to 800 pixels wide: %d", img.Bounds().Dx())
    }
}
