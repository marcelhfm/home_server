package mqtt

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/marcelhfm/home_server/internal/db"
)

func createMessagePubHandler(dbs *sql.DB) mqtt.MessageHandler {
	return func(client mqtt.Client, msg mqtt.Message) {
		fmt.Printf("MQTT: Received message: %s from topic %s\n", msg.Payload(), msg.Topic())

		message := string(msg.Payload())
		parts := strings.Split(message, ",")

		dsId, err1 := strconv.Atoi(parts[0])
		moisture, err2 := strconv.ParseFloat(parts[1], 64)
		if err1 != nil || err2 != nil {
			fmt.Println("MQTT: Error parsing message parts")
			return
		}

		currTimestamp := time.Now().Format(time.RFC3339)
		db.IngestIotData(dbs, dsId, "moisture", int(moisture*100), currTimestamp)
	}
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("MQTT: Connected!")

	// Resubscribe to topics upon reconnect
	if token := client.Subscribe("#", 1, nil); token.Wait() && token.Error() != nil {
		fmt.Printf("MQTT: Subscription error: %v", token.Error())
		return
	}
	fmt.Println("MQTT: Subscribed to all topics")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("MQTT: Connection lost: %v\n", err)
	const maxRetries = 5
	const retryDelay = 5 * time.Second

	for i := 0; i < maxRetries; i++ {
		if token := client.Connect(); token.Wait() && token.Error() == nil {
			fmt.Printf("MQTT: Successfully reconnected to mqtt broker.\n")
			break
		} else {
			fmt.Printf("MQTT: Reconnection error %v\n", token.Error())
			if i < maxRetries-1 {
				fmt.Printf("MQTT: Retrying in %v...\n", retryDelay)
				time.Sleep(retryDelay)
			} else {
				fmt.Println("MQTT: Max retries reached, giving up.")
				return
			}
		}
	}
}

func StartMqttListener(db *sql.DB) {
	var broker = "192.168.11.30"
	var port = 1883
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	opts.SetClientID("home_server")
	opts.SetDefaultPublishHandler(createMessagePubHandler(db))
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts)

	const maxRetries = 5
	const retryDelay = 5 * time.Second

	for i := 0; i < maxRetries; i++ {
		if token := client.Connect(); token.Wait() && token.Error() == nil {
			fmt.Printf("MQTT: Successfully connected to mqtt broker.\n")
			break
		} else {
			fmt.Printf("MQTT: Connection error %v\n", token.Error())
			if i < maxRetries-1 {
				fmt.Printf("MQTT: Retrying in %v...\n", retryDelay)
				time.Sleep(retryDelay)
			} else {
				fmt.Println("MQTT: Max retries reached, giving up.")
				return
			}
		}
	}
}
