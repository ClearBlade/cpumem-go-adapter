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
	disk "github.com/shirou/gopsutil/v3/disk"
	host "github.com/shirou/gopsutil/v3/host"

)

const (
	adapterName = "cpumem-go-adapter"
	topicRoot = "normalizer-generic/data"
	infoTopic = "normalizer-generic/data"
	assetType = "edge"
	groupID = "default"
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
	sendGatewayInfo()
	// kick off adapter specific things here
	go sendSystemStats()
	// keep adapter executing indefinitely 
	select {}
}

func sendGatewayInfo() {
	h, _ := host.Info()
	d, _ := disk.Usage("/")
	c, _ := cpu.Info()

	var gatewayInfoMarshalled []byte
	var err error
	gatewayInfo := map[string]interface{}{
			"os":h.OS,
			"os_platform":h.Platform,
			"os_platform_family":h.PlatformFamily,
			"kernel_arch":h.KernelArch,
			"os_platform_version":h.PlatformVersion,
			"model_name":c[0].ModelName,
			"cache_size":c[0].CacheSize,
			"total_disk_space":d.Total,
			"used_disk_space_percent":d.UsedPercent,
			"is_validated":true,
		}

	asset := map[string]interface{}{
		"type":assetType,
		"custom_data":gatewayInfo,
		"group_id":groupID,
	}

	if gatewayInfoMarshalled, err = json.Marshal(asset); err != nil {
		fmt.Println(err)
		return
	}
	
	fmt.Println(string(gatewayInfoMarshalled))
	adapter_library.Publish(infoTopic, gatewayInfoMarshalled)
}

func sendSystemStats() {
	for {
		time.Sleep(20 * time.Second)
		v, _ := mem.VirtualMemory()
		c, _ := cpu.Percent(10 * time.Second, false)
		d, _ := disk.Usage("/")
	
		var mCM []byte
		var err error
		var virtualMemory map[string]interface{}

		customData := make(map[string]interface{})
		if len(c) != 0 {
			customData["cpu_used_percent"]=c[0]
		}

		mV, _ := json.Marshal(v)
		json.Unmarshal(mV, &virtualMemory)

		for key, value := range virtualMemory { 
			customData[key] = value
		}

		customData["used_disk_space_percent"]=d.UsedPercent
		
		asset := make(map[string]interface{})
		asset["custom_data"]=customData
		asset["type"]=assetType
		asset["group_id"]=groupID
		
		if mCM, err = json.Marshal(asset); err != nil {
			fmt.Println(err)
			return
		}
		
		fmt.Println(string(mCM))
		//fmt.Println("Topic Root: ", adapterConfig.TopicRoot)
		adapter_library.Publish(topicRoot, mCM)
	}
}

func cbMessageHandler(message *mqttTypes.Publish) {
	// process incoming MQTT messages as needed here
	fmt.Println("Incoming Message from ClearBlade:- ", message)
}