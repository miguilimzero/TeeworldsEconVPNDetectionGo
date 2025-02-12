
# Teeworlds VPN Detection & Distributed Banserver (written in Go)

This application connects to teeworlds servers via its configured external console (econ).
It reads every logged line and checks for joining players.
The joining player's IP is then compared to the redis cache.
Does the cache not contain the IP, currently three VPN detection APIs are used to determine whether the player's IP is a VPN or not.
60% of these three APIs need to detect the IP as VPN in order for the application to actually ban the player and cache his VPN IP in the redis cache as such.

## Requirements

### Redis server for caching of IPs

This application requires a running redis database that can be used as cache for IPs.
The application caches non-VPN IPs in the redis database for one week.
VPN IPs are saved forever in order not to hit the free rate limit of the used APIs too fast.

#### Debian & Ubuntu

On Linux it is usually started automatically after its installation.

```shell

sudo apt install redis-server
```

#### macOS

On macOS you need to manually start it with `redis-server`.

```shell
brew install redis
```

## .env configuration file

The `.env` file contains the configuration.
Especially the econ addresses and password of the servers that this application should be attached to.

- .env file in the same folder as the executable

### Example .env

This file needs to live within the same directory as your executable.

```env
# .env

# the api key can be found here: https://proxycheck.io/ (requires registration)
PROXYCHECK_TOKEN=abcdefghijklmnopqrst0123456789

# econ addresses separated by one space
ECON_ADDRESSES=127.0.0.1:9303 127.0.0.1:9304 127.0.0.1:9305

# passwords, either one for all or one per address
ECON_PASSWORDS=abcdef0123456789

# each connection retries incrementally to reconnect to the server.
# if the connection timeout is reached, the routine stops.
RECONNECT_TIMEOUT=24h

# retries to reconnect a lost connection after x seconds
RECONNECT_DELAY=10s

# redis database credentials, these are the default values after installation
REDIS_ADDRESS=localhost:6379
REDIS_PASSWORD=

# how many minutes the VPN IP is banned and with what reason.
# 24h10m, smallest unit are minutes. any fraction of a minute is cut off.
VPN_BANTIME=24h
VPN_BANREASON=VPN - discord.gg/123456

# specific delayed banning that waits for a specific joining state, where all player information is 
# already known to the zCatch server (enables zCatch specific log parsing: "1", "true", "enable", "enabled", "on")
# any other value disables this parsing variant.
ZCATCH_LOGGING=0


```

## Downloading dependencies

```shell
go get -d
```

## Building

```shell
go build .
```

## Running

```shell
./TeeworldsEconVPNDetectionGo
```

## Add/Remove IPs from IPv4 text file to/from the Redis database

In order for this to work, you need to have a properly configured setup with a `.env` file.
Given a file with conents like:

```text
1.236.132.203

# this adds/removes 2.56.0.1 through 2.56.255.254 to the database
2.56.92.0/16

# this adds/removes the IPs 2.56.140.1 through 2.56.140.254 from the database
2.56.140.0/24

123.0.0.1 # add any custom ban reason
123.0.0.1/24 # also add any custom ban reason, # followed by text

213.182.158.200 - 213.182.158.203 # reason (excluding the upper boundary IP, IPs ending with 0 or 255)

```

You can automatically add all of those IPs and the IPs from the subnets to your redis cache.
In order for such a file to be parsed, you can pass it on startup to the application like this:

```text
# add IPs that are supposed to be banned to the database
./TeeworldsEconVPNDetectionGo -add ips.txt

# remove IPs from the database
./TeeworldsEconVPNDetectionGo -remove ips.txt

# whitelist IPs forever in case the utilized APIs provide false positives
./TeeworldsEconVPNDetectionGo -whitelist whitelist.txt

# run the detection in ofline mode. This allows basically to abuse the detection as a banserver.
./TeeworldsEconVPNDetectionGo -offline
```

You can use the `ip-v4.txt` from [VPNs](https://github.com/ejrv/VPNs).
Due to the underlying *goripr* library the insertion of those IP ranges is pretty fast and storage efficient.

After all of the IPs have been parsed and added to the cache, the application shuts down.
You need to restart it without the flag in order to have the econ VPN detection behavior.

## Note

Currently no IPv6 support.  
Add a `# ban reason` behind the IP or behind the IP range to add a custom ban reason.

## Troubleshooting

### EOF error

The Teeworlds server banned you for attempting to login too any times or for connecting too often.
Possible solution is to restart the game server. Should not be an issue, as the detector attempts to reconnect to the server.
