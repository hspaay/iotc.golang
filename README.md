# iotc.golang

iotc.golang is an implementation of the IoTConnect (iotc) standard for publishing and subscribing to IoT information on a message bus. This library is part of the reference implementation. The standard can be found at: https://github.com/hspaay/iotc.standard

## Audience

This library is intended for software developers that wish to develop IoT applications in the golang language. A similar library for the Python and Javascript/Typescript languages is planned for the future.

## This Library Provides

* systemd launcher of adapters for Linux 
* Messengers MQTT brokers and testing (DummyMessenger)
* Management of nodes, inputs and outputs (see IoTConnect standard for further explanation)
* Publish discovery when nodes are updated
* Publish updates to output values
* Signing of published messages
* Hook to handle node input control messages
* Hook to handle node configuration updates
* Constants and Type Definitions of the IoTConnect standard

## Prerequisites

1. Golang
   
   This guide assumes that you are familiar with programming in golang, and have golang 1.13 or newer installed. If you are new to golang, check out their website https://golang.org/doc/install for more information. 

2. MQTT broker
  
   A working [MQTT broker](https://en.wikipedia.org/wiki/MQTT) (message bus) is needed to test and run a publisher. [Mosquitto](https://mosquitto.org/) is a lightweight MQTT broker that runs nicely on Linux (including Raspberry-pi), Mac, and Windows. Only a single broker is needed for all your publishers. For a home automation application you will do fine with running Mosquitto on a Raspberry-pi 2, 3 or 4 connected to a small UPS and park it somewhere out of sight.

   For industrial or government applications that require GDPR or SOC2 compliance, look at enterprise message brokers such as [HiveMQT](www.hivemq.com), [RabbitMQT](https://www.rabbitmq.com/), [Apache ActiveMQ](https://activemq.apache.org/). For hosted cloud versions look at [CloudMQTT](www.cloudmqtt.com) (uses managed Mosquitto instances under the hood). Amazon AWS and Google also support IoT message buses and are worth a look.

   See the section Installing Mosquitto for use of mosquitto.

3. Access to git and github.com is needed to retrieve the development libraries. 

4. A code editor or IDE is needed to edit source code. Visual Studio Code is a free IDE that is quite popular and has good support for golang. See https://code.visualstudio.com/ for more information and downloads.

5. Publisher specific dependencies. These are listed in the publisher's README.md file.


## How To Use This Library

under construction...

for now, see the [EXAMPLE.md]
API docs are found under docs


## Building and Installing Publishers

Publishers can be build from source, or an x64 or arms binary can simply be downloaded from the releases folder. 

The publisher can be installed in a personal bin folder (~/bin/iotc/bin), or system wide (/usr/local/bin) along with a systemd startup file.

### Build From Source 

Clone the source from the git repository and invoke go build. Some publishers might require additional libraries to be installed, which are listed in the publisher's README.md.

This uses the iotc.ipcam publisher as an example

```bash
git clone https://github.com/hspaay/iotc.ipcam
cd iotc.ipcam
go build
```
This results in a binary file 'iotc.ipcam'. The folder test/ contains example messenger.yaml and ipcam.yaml configuration files.

### Download Release Files

Binaries for x64 and arm, along with default configuration files can be downloaded from the releases folder (using iotc.ipcam as an example).

Open https://github.com/hspaay/iotc.ipcam/releases in a web browser and download the latest release executable iotc.ipcam for your platform (x86_64 or arm), and the default configuration file ipcam.yaml.

The https://github.com/hspaay/iotc.golang/releases contains a sample messenger.yaml which is needed to setup the connection to the MQTT broker. 


### Local User Installation (Linux)

Local user installation doesn't require root or sudo permissions. The folder structure for deployment as a normal user:
* ~/bin/iotc/bin      location of the publisher binaries
* ~/bin/iotc/config   location of the configuration files, including iotc.conf
* ~/bin/iotc/logs     logging output

Install locally (using iotc.ipcam as an example):
```bash
mkdir -P ~/bin/iotc/bin
mkdir -P ~/bin/iotc/logs
mkdir -P ~/bin/iotc/config

chmod +x iotc.ipcam
cp iotc.ipcam ~/bin/iotc/bin
cp -n test/messenger.yaml test/ipcam.yaml ~/bin/iotc/config/
```

Edit messenger.yaml with the correct mqtt server address, login credentials and zone. The zone is only required if you want to share information with other zones. It can be any name that is unique amongst other zones. For global world sharing the zone has to be globally unique, eg like a domain name. This file is shared amongst all publishers and only needs to be configured once.

Edit ipcam.yaml configuration file. See the iotc.ipcam README for details. Many publishers support a quick start configuration using the configuration file and support more extensive configuration using the publisher and node configuration messages. This requires a iotc compatible UI.

Add ~/bin/iotc/bin to your PATH in ~/.bashrc (don't forget to open another shell to activate the change)

To run:

> iotc.ipcam

### System Installation (Linux)

This requires root or sudo permissions. 

Iotc publishers run as a systemd service, by default using the 'iotc' user and group.
The startup configuration files can be found in /etc/iotc/. Systemd is used for autostart. Edit iotc.target to include the services to autostart.

When deploying as an system application, create these folders and update /etc/iotc/messenger.yaml
* /etc/iotc/conf         location of iotc.conf main configuration file
* /opt/iotc/             location of the publisher binaries
* /var/lib/iotc/         location of the persistence files
* /var/log/iotc/         location of iotc log files


Step 1: Setup system for running iotc publishers. This is only needed once.
```bash
sudo useradd iotc
sudo mkdir -P /var/log/iotc/
sudo mkdir -P /etc/iotc/ 
sudo chmod 750 /etc/iotc/
sudo chmod 750 /var/log/iotc/
sudo chown iotc /etc/iotc/
sudo chown iotc /var/log/iotc/
sudo cp -n messenger.yaml /etc/iotc/
sudo cp iotc.target /etc/systemd/system
sudo cp iotc@.service /etc/systemd/system
sudo cp iotc.logrotate /etc/logrotate/logrotate.d/iotc
sudo systemctl enable iotc.target
sudo systemctl start iotc.target
```

Step 2: Install the publisher(s). This is done for each publisher.
```bash
chmod +x iotc.ipcam
sudo cp iotc.ipcam /usr/local/bin
sudo cp -n test/ipcam.yaml /etc/iotc/
```

## Install Mosquitto

[Mosquitto](https://mosquitto.org/) is a lightweight MQTT server and a great option for use as the IoTConnect message bus. Installation for the different platforms[is described here](https://mosquitto.org/download/).

Mosquitto needs very little configuration as [described on their documentation](https://mosquitto.org/man/mosquitto-conf-5.html).
On Linux the configuration takes place in the /etc/mosquitto/mosquitto.conf file.

### Configuration

The configuration defaults are good to go for use on a trusted LAN. For external exposure you're going to want to add client authentication.

The ZCAS service will make life easier by auto-configuring mosquitto and the publishers to run securely. This is currently work in progress. See the [iotc.zcas https://github.com/hspaay/zcas] publisher for details.

### Automatically Start Mosquitto On Boot [Linux]

On Linux:
$ sudo systemctl enable mosquitto
$ sudo systemctl start mosquitto



## Contributing

Contributions to the IoTConnect project are very welcome. There are many areas where help is needed, especially with documentation and building publishers for IoT and other devices.
See [CONTRIBUTING](docs/CONTRIBUTING.md) for guidelines.

## Questions

For issues and questions, please open a ticket.
Common questions will be captured in the [Q&A](docs/FAQ.md).


