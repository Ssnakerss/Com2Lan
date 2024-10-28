package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	channelName = "broadcast"
)

func fail(err error, message string) {
	if err != nil {
		log.Fatal(message, "err:", err)
	}
}

func main() {

	mctx, mcancel := context.WithCancel(context.Background())
	defer mcancel()

	go func() {
		exit := make(chan os.Signal, 1)
		signal.Notify(exit,
			syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT)
		sig := <-exit
		log.Printf("signal received %v", sig)
		log.Printf("stopping server")
		mcancel()
	}()

	go consumer(mctx, "consumer 1") //consumer 1
	go consumer(mctx, "consumer 2") //consumer 1
	go consumer(mctx, "consumer 3") //consumer 1

	//---------------------------------------------
	// producer

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	fail(err, "dial")
	defer func() { conn.Close() }()

	ch, err := conn.Channel()
	fail(err, "channel")
	defer func() { ch.Close() }()

	err = ch.ExchangeDeclare(
		channelName, //name
		"fanout",    //type
		true,        //durable
		false,       //autoDelete
		false,       //internal
		false,       //noWait
		nil,
	)
	fail(err, "exchange")

	body := "Hello World!"
	for i := 0; i < 1000_000; i++ {
		select {
		case <-mctx.Done():
			ch.ExchangeDelete(channelName, false, false)
			return
		default:
			time.Sleep(time.Second * 3)
			err = ch.Publish(
				"broadcast", //exchange
				"",          //routing key
				false,       //mandatory
				false,       //immediate
				amqp.Publishing{
					Headers: amqp.Table{
						"ContentType": "text/plain",
						"sender":      "me is sender",
					},
					ContentType: "text/plain",
					ReplyTo:     "auto sender !!!", //user
					Body:        []byte(body + strconv.Itoa(i)),
				})
			fail(err, "publish")
			log.Printf("  [*] Sent a message %s%d", body, i)
		}
	}
}

// consumer
func consumer(ctx context.Context, name string) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	fail(err, name+":dial")
	defer func() { conn.Close() }()

	ch, err := conn.Channel()
	fail(err, name+":channel")
	defer func() { ch.Close() }()

	q, err := ch.QueueDeclare(
		"",    //name если не указано - генерится автоматом
		false, //durable
		false, //delete when unused
		true,  //exclusive
		false, //no-wait
		nil,   //arguments
	)
	fail(err, name+":queue")

	err = ch.QueueBind(
		q.Name,
		"",
		channelName,
		false,
		nil,
	)
	fail(err, name+":queue bind")

	//---------------------------------------------------------
	messages, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	fail(err, name+":consumer")

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case message := <-messages:
				log.Printf("  [*][%s] received a message : %s from %v", name, message.Body, message.Headers["sender"])
			}
		}
	}()

	log.Printf("  [*][%s] Waiting for messages. ", name)
	<-ctx.Done()
	log.Printf("  [*][%s] Complete", name)
	return
}
