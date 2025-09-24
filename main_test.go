package main

import (
	"reflect"
	"testing"
)

func TestParseBluetoothOutput(t *testing.T) {
	sampleOutput := `Bluetooth:

      Bluetooth Controller:
          Address: 1C:1D:D3:D7:16:DB
          State: On
          Chipset: BCM_4388C2
          Discoverable: Off
          Firmware Version: 22.5.190.749
          Product ID: 0x4A36
          Supported services: 0x392039 < HFP AVRCP A2DP HID Braille LEA AACP GATT SerialPort >
          Transport: PCIe
          Vendor ID: 0x004C (Apple)
      Not Connected:
          ChrisiPhone16:
              Address: 6C:3A:FF:C7:94:90
              Vendor ID: 0x004C
              Product ID: 0x2014
              Minor Type: Phone
              RSSI: -60
          Chris's Daily Watch:
              Address: 94:0B:CD:1F:BB:1B
              Vendor ID: 0x004C
              Product ID: 0x1234
              Minor Type: Watch
              RSSI: -76
      Connected:
          AirPods Pro:
              Address: 2C:76:00:D9:9E:58
              Vendor ID: 0x004C
              Product ID: 0x2014
              Minor Type: Headphones
              RSSI: -45`

	devices, err := parseBluetoothOutput(sampleOutput)
	if err != nil {
		t.Fatalf("parseBluetoothOutput() error = %v", err)
	}

	expected := []BluetoothDevice{
		{Name: "ChrisiPhone16", Address: "6C:3A:FF:C7:94:90", RSSI: -60},
		{Name: "Chris's Daily Watch", Address: "94:0B:CD:1F:BB:1B", RSSI: -76},
		{Name: "AirPods Pro", Address: "2C:76:00:D9:9E:58", RSSI: -45},
	}

	if len(devices) != len(expected) {
		t.Fatalf("Expected %d devices, got %d", len(expected), len(devices))
	}

	for i, device := range devices {
		if !reflect.DeepEqual(device, expected[i]) {
			t.Errorf("Device %d: expected %+v, got %+v", i, expected[i], device)
		}
	}
}

func TestParseBluetoothOutputNoRSSI(t *testing.T) {
	sampleOutput := `Bluetooth:

      Bluetooth Controller:
          Address: 1C:1D:D3:D7:16:DB
          State: On
      Not Connected:
          Some Device:
              Address: 6C:3A:FF:C7:94:90
              Vendor ID: 0x004C
              Product ID: 0x2014`

	devices, err := parseBluetoothOutput(sampleOutput)
	if err != nil {
		t.Fatalf("parseBluetoothOutput() error = %v", err)
	}

	if len(devices) != 0 {
		t.Errorf("Expected 0 devices (no RSSI), got %d", len(devices))
	}
}

func TestParseBluetoothOutputEmptyString(t *testing.T) {
	devices, err := parseBluetoothOutput("")
	if err != nil {
		t.Fatalf("parseBluetoothOutput() error = %v", err)
	}

	if len(devices) != 0 {
		t.Errorf("Expected 0 devices for empty output, got %d", len(devices))
	}
}

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	configContent := `mqtt:
  server: "test.mosquitto.org"
  port: 1883
  username: "testuser"
  password: "testpass"
  root_topic: "test/bluetooth"`

	tmpfile, err := createTempFile("config_test.yaml", configContent)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = removeTempFile(tmpfile) }()

	config, err := loadConfig(tmpfile)
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}

	if config.MQTT.Server != "test.mosquitto.org" {
		t.Errorf("Expected server 'test.mosquitto.org', got '%s'", config.MQTT.Server)
	}
	if config.MQTT.Port != 1883 {
		t.Errorf("Expected port 1883, got %d", config.MQTT.Port)
	}
	if config.MQTT.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", config.MQTT.Username)
	}
	if config.MQTT.RootTopic != "test/bluetooth" {
		t.Errorf("Expected root_topic 'test/bluetooth', got '%s'", config.MQTT.RootTopic)
	}
}

func TestLoadConfigInvalidFile(t *testing.T) {
	_, err := loadConfig("nonexistent.yaml")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}