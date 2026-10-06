# Config Example

The setup for this example is a midi controller (AKAI LPD8 or similar) to send MIDI messages to the windows device.

``` txt
[AKAI LPD8] -- midi --> GoDing --+--> System volume
                                 +--> Application volume
                                 +--> Output device volume
```

The configuration example acts on defined MIDI messages to control the volume of the
system sounds, an application output volume and the volume of the current output device.

``` JSON
{
  "Element": {
    "e1": {
      "key1": "value1",
      "key2": "value2"
    }
  },
  "Volume": {
    "main": {
      "min": "0",
      "max": "100",
      "endpoint": "out:con",
      "value": "50"
    },
    "bee": {
      "min": "0",
      "max": "100",
      "endpoint": "app:bee",
      "value": "20"
    },
    "syssound": {
      "min": "0",
      "max": "100",
      "endpoint": "app:system",
      "value": "80"
    }
  },
  "midi": {
    "K5": {
      "message": "[14] CC 74",
      "onMessage": "volume/syssound?value=$v"
    },
    "K7": {
      "message": "[14] CC 76",
      "onMessage": "volume/bee?value=$v"
    },
    "K8": {
      "message": "[14] CC 77",
      "onMessage": "volume/main?value=$v"
    }
  }
}
```