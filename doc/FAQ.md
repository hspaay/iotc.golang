# IoTConnect Q&A

## Is this yet another home automation project? 
Projects like [Home Assistant](https://www.home-assistant.io/) and [OpenHab](https://www.openhab.org/) aim to provide a full featured home automation solution. If this is what you are interested in, please do check out those projects. 

IoTConnect aims to collect and exchange information from IoT devices and related services. Information from IoT devices can easily be integrated into any project that supports the MQTT message bus. It complements existing home automation systems by providing sensor data.


## Does this duplicate 'bindings' from automation projects?
Unfortunately, this can lead to duplication of bindings implementation. In IoTConnect these are called *publishers*. 

Among IoTConnect's benefits are that the information published is self descriptive, supports discovery, and supports configuration of the nodes that provide the information. Publishers can be implemented in any programming language and are therefore not tied to the implementation language of the home automation solution. 

Most importantly though is that 'bindings' implemented following the IotConnect standard are interoperable with many automation solutions that support the MQTT message bus. This benefits all automation projects rather than just a single one. The IotConnect standard is not restricted to the use of MQTT. Any publish-subscribe message bus can be used, although it does require a messenger library for the new message bus.


## How does IoTConnect differ from home automation projects?
IoTConnect focuses on collecting and publishing of information as an open standard, for any purpose. Any IoT sensor network can use it to aggregate and present information. This can be for an industrial application, governmental use, non profit services, or personal projects.

Home automation projects integrate with IoTConnect through support of the MQTT message bus. It can even be used as the complete data aggregation layer in existing home automation projects if desired.

Sensors that are not 'internet connected' can easily be made accessible by combining them with a small computer board with I/O interfaces, such as the Raspberry-Pi or industrial equivalent. A 'publisher' can be written for these computers that translates between the raw I/O data and the IotConnect standard. From there on the information can be used like any other information that is published following the standard.


## Does IoTConnect require the use of the MQTT message bus.
The IotConnect standard does not specify a particular message bus as transport. In fact, it can be implemented using a REST API, AMQP, or the Microsoft message bus as transports. A so-called 'bridge' can be used to connect between zones that use different transports.

The reference implementation uses the MQTT message bus as this is lightweight, suitable to IoT devices due to its low overhead, and has wide industry support.
