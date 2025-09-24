package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"gopkg.in/yaml.v3"
)

type Config struct {
	MQTT struct {
		Server    string `yaml:"server"`
		Port      int    `yaml:"port"`
		Username  string `yaml:"username"`
		Password  string `yaml:"password"`
		RootTopic string `yaml:"root_topic"`
	} `yaml:"mqtt"`
}

type BluetoothDevice struct {
	Name    string
	Address string
	RSSI    int
}

func loadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

func scanBluetoothDevices() ([]BluetoothDevice, error) {
	cmd := exec.Command("system_profiler", "SPBluetoothDataType")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run system_profiler: %w", err)
	}

	return parseBluetoothOutput(string(output))
}

func parseBluetoothOutput(output string) ([]BluetoothDevice, error) {
	devices := []BluetoothDevice{}
	lines := strings.Split(output, "\n")

	var currentDevice string
	var currentAddress string

	for _, line := range lines {
		// Match device names (indented lines ending with colon)
		if matched, _ := regexp.MatchString(`^[[:space:]]{10}[^:]+:$`, line); matched {
			// Skip system entries
			if !strings.Contains(line, "Bluetooth") && !strings.Contains(line, "Controller") &&
			   !strings.Contains(line, "Features") && !strings.Contains(line, "Services") &&
			   !strings.Contains(line, "Connected") && !strings.Contains(line, "Not Connected") {
				currentDevice = strings.TrimSpace(strings.TrimSuffix(line, ":"))
				currentAddress = ""
			}
		}

		// Extract MAC address
		if strings.Contains(line, "Address:") && currentDevice != "" {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				addrPart := strings.TrimSpace(strings.Join(parts[1:], ":"))
				if matched, _ := regexp.MatchString(`[0-9A-Fa-f:]+`, addrPart); matched {
					re := regexp.MustCompile(`[0-9A-Fa-f:]+`)
					if match := re.FindString(addrPart); match != "" {
						currentAddress = match
					}
				}
			}
		}

		// Extract RSSI value
		if strings.Contains(line, "RSSI:") && currentDevice != "" {
			re := regexp.MustCompile(`-?[0-9]+`)
			if match := re.FindString(line); match != "" {
				if rssi, err := strconv.Atoi(match); err == nil {
					address := currentAddress
					if address == "" {
						address = "Unknown"
					}
					devices = append(devices, BluetoothDevice{
						Name:    currentDevice,
						Address: address,
						RSSI:    rssi,
					})
					currentDevice = ""
					currentAddress = ""
				}
			}
		}
	}

	return devices, nil
}

func publishToMQTT(config *Config, devices []BluetoothDevice) error {
	brokerURL := fmt.Sprintf("tcp://%s:%d", config.MQTT.Server, config.MQTT.Port)

	opts := mqtt.NewClientOptions().AddBroker(brokerURL)
	if config.MQTT.Username != "" {
		opts.SetUsername(config.MQTT.Username)
	}
	if config.MQTT.Password != "" {
		opts.SetPassword(config.MQTT.Password)
	}
	opts.SetClientID("bt-env-mqtt")
	opts.SetConnectTimeout(10 * time.Second)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
	}
	defer client.Disconnect(250)

	for _, device := range devices {
		// Publish RSSI
		rssiTopic := fmt.Sprintf("%s/%s/rssi", config.MQTT.RootTopic, device.Address)
		rssiValue := fmt.Sprintf("%d", device.RSSI)
		if token := client.Publish(rssiTopic, 0, false, rssiValue); token.Wait() && token.Error() != nil {
			return fmt.Errorf("failed to publish RSSI for device %s: %w", device.Address, token.Error())
		}

		// Publish device name
		nameTopic := fmt.Sprintf("%s/%s/name", config.MQTT.RootTopic, device.Address)
		if token := client.Publish(nameTopic, 0, false, device.Name); token.Wait() && token.Error() != nil {
			return fmt.Errorf("failed to publish name for device %s: %w", device.Address, token.Error())
		}
	}

	return nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <config-file>\n", os.Args[0])
		os.Exit(1)
	}

	configFile := os.Args[1]

	config, err := loadConfig(configFile)
	if err != nil {
		log.Printf("Error loading config: %v", err)
		os.Exit(1)
	}

	devices, err := scanBluetoothDevices()
	if err != nil {
		log.Printf("Error scanning Bluetooth devices: %v", err)
		os.Exit(1)
	}

	if len(devices) == 0 {
		log.Printf("No Bluetooth devices found with RSSI data")
		os.Exit(0)
	}

	if err := publishToMQTT(config, devices); err != nil {
		log.Printf("Error publishing to MQTT: %v", err)
		os.Exit(1)
	}

	log.Printf("Successfully published data for %d devices", len(devices))
	os.Exit(0)
}