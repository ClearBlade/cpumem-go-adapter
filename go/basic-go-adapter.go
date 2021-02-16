package main

import (
	"log"
	"encoding/json"
	"fmt"
	"time"
	adapter_library "github.com/clearblade/adapter-go-library"
	mqttTypes "github.com/clearblade/mqtt_parsing"
	mem "github.com/shirou/gopsutil/v3/mem"
	cpu "github.com/shirou/gopsutil/v3/cpu"

)

const (
	adapterName = "cpumem-go-adapter"
	topicRoot = "_monitor/asset/gateway-validation/data"
)

var (
	adapterConfig    *adapter_library.AdapterConfig
)

func main() {

	// add any adapter specific command line flags needed here, before calling ParseArguments
	err := adapter_library.ParseArguments(adapterName)
	if err != nil {
		log.Fatalf("[FATAL] Failed to parse arguments: %s\n", err.Error())
	}

	// initialize all things ClearBlade, includes authenticating if needed, and fetching the
	// relevant adapter_config collection entry
	adapterConfig, err = adapter_library.Initialize()
	if err != nil {
		log.Fatalf("[FATAL] Failed to initialize: %s\n", err.Error())
	}
	
	// if your adapter config includes custom adapter settings, parse/validate them here
	
	// connect MQTT, if your adapter needs to subscribe to a topic, provide it as the first
	// parameter, and a callback for when messages are received. if no need to subscribe,
	// simply provide an empty string and nil
	err = adapter_library.ConnectMQTT(topicRoot+"/outgoing/#", cbMessageHandler)
	if err != nil {
		log.Fatalf("[FATAL] Failed to Connect MQTT: %s\n", err.Error())
	}
	
	// kick off adapter specific things here
	go sendSystemStats()
	// keep adapter executing indefinitely 
	select {}
}

func sendSystemStats() {
	for {
		v, _ := mem.VirtualMemory()
		c, _ := cpu.Percent(10 * time.Second, false)
		
		var mCM []byte
		var err error
		// if mV, err = json.Marshal(v); err != nil {
		// 	fmt.Println(err)
		// 	return
		// }
		
		cpuMem := map[string]interface{}{
			"mem": v,
			"cpu":c[0],
		}
		
		if mCM, err = json.Marshal(cpuMem); err != nil {
			fmt.Println(err)
			return
		}
		
		fmt.Println(string(mCM))
		//fmt.Println("Topic Root: ", adapterConfig.TopicRoot)
		adapter_library.Publish(topicRoot, mCM)
		time.Sleep(10 * time.Second)
	}
}

func cbMessageHandler(message *mqttTypes.Publish) {
	// process incoming MQTT messages as needed here
	fmt.Println("Incoming Message from ClearBlade: ", message)
}