## Architecture

```txt

                    +-------------------+
                    |     Sensor(s)     |
                    |-------------------|
                    |      LoraWan      |
                    +-------------------+
                              |
                              |
                           LoraWan
                              |
                              v
                    +-------------------+
                    |      Gateway      |
                    |-------------------|
                    | - LoraWan         |
                    | - MQTT Publisher  |
                    | - SQLite          |
                    |  (ML Prediction)  |
                    +-------------------+
                              |
                              |
                         MQTT Publish
                              |
                              v
                  +-------------------------+
                  |       MQTT Broker       |
                  |-------------------------|
                  | (EMQX, Mosquitto, etc.) |
                  +-------------------------+
                              ^
                              |
                        Internet (TLS)
                              |
                        MQTT Subscribe
                              |
                              |
                +-----------------------------+
                |      Consumer (Cloud)       |
                |-----------------------------|
                | - MQTT Subscriber           |
                | - Lưu Data vào SQLite       |
                | - Websocket Server          |
                | - HTTP Server (HTMX + UI)   |
                | - Tailwind UI Templates     |
                +-----------------------------+
                              |
                              |
             +----------------+----------------+
             |                                 |
  WebSocket (Push realtime)            HTTP Query (HTMX)
             |                                 |
   +--------------------+           +-------------------------+
   | Realtime Live View |           | Historical Query + Alert|
   +--------------------+           +-------------------------+
```
