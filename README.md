# MyDNS

MyDNS is a DNS server application written in Go. It is designed to be simple, efficient, and easy to use.

## Features

- [x] Lightweight and fast
- [x] Easy to configure
- [x] Basic DNS functionalities: query and block domains
- [ ] Load balancing

## Requirements

- Go 1.16 or later
- Redis 6.0 or later (optional)

## Installation

1. Clone the repository:
    ```sh
    git clone https://github.com/darwineee/MyDNS.git
    cd mydns
    ```

2. Build the application:
    ```sh
    make build
    ```
   The executable file will be generated in the `bin` directory.

## Usage

1. Create a configuration file named `config.yaml` in the root directory of the application. Available options are
   demonstrated in source code
   `config.yaml` file as default values. If you want to use the default values, you can skip this step.

2. Create `blacklist` and `known_hosts` files in the root directory of the application. The `blacklist` file contains
   a list of domain names only, whereas `known_hosts` contains a map of domain names and IP addresses. Put each entry
   one per line.
   The `blacklist` file can be used to block access to certain domains, while the `known_hosts` file can be used to
   resolve them.
   Both files are optional. Skip this step if you don't need them.

3. Run the application:
    ```sh
    make run
    ```

4. The application will start and listen for DNS queries on port 2053 using UDP by default.
   You can change these settings in the configuration file. However, in practice, you don't need to change them.
   ![Application banner](/screenshots/application_started_normal.png)

5. To stop the application, press `Ctrl + C` or type the command `stop`. More commands will be supported in the future.

6. This application uses the system resolver. You can add more foreign name servers to the OS's resolver configuration
   file.
   For example, it is `/etc/resolv.conf` on Unix-like systems.

## Development

This application implements the [RFC-1035](https://www.rfc-editor.org/rfc/rfc1035) specification.

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

Redis is used as a caching layer.

To handle high QPS, I implemented an event loop to handle incoming DNS queries concurrently. The event loop is as simple
as below:
![Event loop](/screenshots/event_loop_design.png)