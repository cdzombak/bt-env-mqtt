# bt-env-mqtt

A macOS tool that scans for nearby Bluetooth devices and publishes their RSSI (signal strength) and names to MQTT.

## Installation

TK

## Usage

```bash
./bt-env-mqtt config.yaml
```

## Configuration

Create a YAML configuration file (see `config.yaml.example`):

```yaml
mqtt:
  server: "localhost"
  port: 1883
  username: ""
  password: ""
  root_topic: "my/computer/bluetooth-env"
```

## MQTT Topics

For each discovered Bluetooth device, the tool publishes to:

- `<root topic>/<device MAC address>/rssi` - RSSI value (signal strength)
- `<root topic>/<device MAC address>/name` - Device name

Example topics:
- `my/computer/bluetooth-env/6C:3A:FF:47:74:9B/rssi` → `-60`
- `my/computer/bluetooth-env/6C:3A:FF:47:74:9B/name` → `ChrisiPhone16`

## Home Assistant

To pull a single device's RSSI into Home Assistant (for example, for use in presence detection), add to your `configuration.yaml`:

```yaml
  sensor:
    - name: "iPhone RSSI"
      unique_id: "mycomputer/iphone_rssi"
      device:
        name: "My Computer"
        identifiers:
        - "mycomputer"
      state_topic: "dzhome/computer/mycomputer/bluetooth/environment/6C:3A:FF:47:74:9B/rssi"
      unit_of_measurement: "dB"
```

## Launchd Integration

To run this as a global daemon on macOS:

1. Modify the paths & other details in [`com.dzombak.bluetooth-env-mqtt.example.plist`](com.dzombak.bluetooth-env-mqtt.example.plist)
2. Write it to `/Library/LaunchDaemons/com.dzombak.bluetooth-env-mqtt.plist`.
3. `sudo chown root:root /Library/LaunchDaemons/com.dzombak.bluetooth-env-mqtt.plist`
4. `sudo launchctl load -w /Library/LaunchDaemons/com.dzombak.bluetooth-env-mqtt.plist`

## License

GNU GPL v3; see [`LICENSE`](LICENSE) in this repo.

## Author

[Chris Dzombak](https://www.dzombak.com) ([GitHub: @cdzombak](https://www.github.com/cdzombak))
