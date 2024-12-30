# MyDNS

MyDNS is a lightweight DNS server implementation written in Go, following the RFC-1035 specification. 
It features an efficient event-loop architecture for handling high query loads and supports basic DNS functionalities with domain blocking capabilities.

## Features

* **High Performance**: Implements an event-loop architecture with worker pools for concurrent DNS query processing
* **DNS Query Support**: Currently handles A record queries, supporting both IPv4 and IPv6 responses
* **Domain Management**:
   * Block unwanted domains using blacklist
   * Custom domain resolution via known_hosts configuration
* **Caching Layer**: Optional Redis integration for improved response times
* **Cross-Platform**: Thanks to Go compiler, supports multiple platforms including:
   * Linux (amd64, arm64)
   * macOS (amd64)
   * Windows (amd64)
* **Load Balancing**: [Planned] Round-robin distribution of DNS queries across worker pools

## Requirements

* Go 1.16 or later
* Redis 6.0 or later (optional)

## Installation

1. Clone the repository:
```bash
git clone https://github.com/darwineee/MyDNS.git
cd mydns
```

2. Build the application:
```bash
# Build for current platform
make build

# Build for all supported platforms
make build-all
```

The executable files will be generated in the `bin` directory.

## Configuration

### Config File (config.yaml)
Create a `config.yaml` file in the root directory. Example configuration:

```yaml
# Redis configuration
redis:
  host: "127.0.0.1:6379"
  password: "admin"
  db: 0

# UDP configuration
udp:
  pkg_limit_rfc1035: 512
  pkg_limit_edns0: 4096

# Server configuration
server:
  port: 2053
  protocol: "udp"
  event_queue_size: 1000
  event_queue_timeout_milliseconds: 500
  cache_ttl_seconds: 300
  blacklist_file_path: "blacklist-example"
  known_hosts_file_path: "known_hosts-example"
```

### Known Hosts File
Create a `known_hosts` file to define custom domain resolutions. Example:

```
# Format: domain IP_address
# Example:
darwindev.direkt.app 51.79.147.45
```

### Blacklist File
Create a `blacklist` file to block specific domains. Example:

```
# One domain per line
blocked-domain.com
unwanted-site.net
```

## Usage

1. Start the server:
```bash
make run
```

2. The server will listen on port 2053 (UDP) by default.

3. To stop the server, press `Ctrl + C` or type `stop` (currently the only supported command).

## Testing

You can test the DNS server using the `dig` command on Linux:

```bash
# Test a domain resolution
dig @localhost -p 2053 example.com A

# Test a known host
dig @localhost -p 2053 darwindev.direkt.app A

# Test a blacklisted domain
dig @localhost -p 2053 blocked-domain.com A
```

## Planned Features

* **Multiple IP Support**: Allow multiple IP addresses per domain in known_hosts file
  ```
  # Example future format:
  darwindev.direkt.app 51.79.147.45 123.222.34.1 1.1.1.1
  ```

* **Extended Query Types**: Support for additional DNS query types beyond A records

* **Command Interface**: Expanded command support beyond 'stop'
   - Status monitoring
   - Configuration reload
   - Cache management
   - Statistics viewing

* **Load Balancing**: Round-robin distribution for domains with multiple IPs

## Architecture

```
                           Local Host          |    Foreign
                                              |
+---------+               +----------+         |  +--------+
|         | user queries  |          |queries  |  |        |
|  User   |-------------->|          |---------|->| Foreign|
| Program |               |  MyDNS   |         |  |  Name  |
|         |<--------------|          |<--------|--| Server |
|         | user responses|          |responses|  |        |
+---------+               +----------+         |  +--------+
                            |     A            |
            cache additions |     | references |
                            V     |            |
                          +----------+         |
                          |  Redis   |         |
                          +----------+         |
```

The application implements a high-performance event loop design:
- Request Handler processes incoming DNS queries
- Queue system manages query distribution
- Worker pool handles concurrent query processing
- Optional Redis caching layer for improved performance

![Event loop](/screenshots/event_loop_design.png)

## Development

This project implements the DNS protocol according to RFC-1035 specifications. The main components include:
- DNS message parsing and encoding
- Query processing and resolution
- Cache management
- Concurrent request handling

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.