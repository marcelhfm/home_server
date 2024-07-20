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
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("MQTT: Connection lost: %v\n", err)
}

func StartMqttListener(db *sql.DB) {
	var broker = "192.168.11.30"
	var port = 1883
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	opts.SetClientID("home_server")
	opts.SetUsername("home_server")
	opts.SetPassword("xxx")
	opts.SetDefaultPublishHandler(createMessagePubHandler(db))
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Printf("MQTT: Connection error %v\n", token.Error())
		return
	}
	fmt.Printf("MQTT: Successfully connected to mqtt broker.\n")

	if token := client.Subscribe("#", 1, nil); token.Wait() && token.Error() != nil {
		fmt.Printf("MQTT: Subscription error: %v", token.Error())
		return
	}

	fmt.Println("MQTT: Subscribed to all topics")
}
