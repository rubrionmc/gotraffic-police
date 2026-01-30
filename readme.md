<a id="readme-top"></a>

<!-- PROJECT SHIELDS -->

[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![License][license-shield]][license-url]
[![GitHub Stars][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]

<br />
<div align="center">
  <a href="https://github.com/rubrionmc/gogate">
    <img src="images/logo.png" alt="Logo" width="80" height="80">
  </a>

  <h3 align="center">Go Traffic Police</h3>

  <p align="center">
    A simple yet powerful lightweight Go gateway for handling and routing TCP traffic.
    <br />
    <a href="https://github.com/rubrionmc/gogate/tree/main/docs"><strong>Explore the docs »</strong></a>
    <br />
    <br />
    <a href="https://rubrion.net/docs/howto/join">View Demo</a>
    &middot;
    <a href="https://github.com/rubrionmc/gogate/issues/new?labels=bug&template=bug-report.md">Report Bug</a>
    &middot;
    <a href="https://github.com/rubrionmc/gogate/issues/new?labels=enhancement&template=feature-request.md">Request Feature</a>
  </p>
</div>

## Table of Contents

<details>
  <summary>Contents</summary>
  <ol>
    <li><a href="#about-the-project">About The Project</a></li>
    <li><a href="#getting-started">Getting Started</a></li>
    <li><a href="#usage">Usage</a></li>
    <li><a href="#roadmap">Roadmap</a></li>
    <li><a href="#contributing">Contributing</a></li>
    <li><a href="#license">License</a></li>
    <li><a href="#contact">Contact</a></li>
  </ol>
</details>

## About The Project

**Go Traffic Police** is a lightweight Go TCP proxy/gateway designed to route traffic efficiently with health checks, configurable timeouts, and easy backend management.

It is built to be simple, minimal, and fully configurable via a TOML file.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Getting Started

Follow these steps to set up the proxy locally.

### Prerequisites

* Go 1.21+ installed
* Git



### Installation

1. Clone the repository:

```sh
git clone https://github.com/rubrionmc/gogate.git
cd gogate
```

2. Build the proxy:

```sh
go build -o gogate .
```

3. Create a `config.toml` (example in `docs/config.example.toml`) or use `-c` to specify a custom path.

```toml
[server]
listen = ":25565"

[timeouts]
backend_dial = "5s"
healthcheck_dial = "3s"
healthcheck_interval = "5s"

[backends]
servers = ["localhost:9001", "localhost:9002"]
```

4. Run the proxy:

```sh
./gogate           # uses ./config.toml by default
./gogate -c prod.toml  # specify custom config
```

<p align="right">(<a href="#readme-top">back to top</a>)</p>



## Usage

* Proxy automatically routes clients to healthy backends.
* Health checks run on configurable intervals.
* All parameters (timeouts, listen address, backends) are configurable via TOML.

Example:

```sh
./gogate -c config.toml
```

* Logs show connection activity and backend health status.

<p align="right">(<a href="#readme-top">back to top</a>)</p>



## Roadmap

* [x] Configurable via TOML
* [x] Health checks for backends
* [ ] TLS support
* [ ] Metrics endpoint
* [ ] Round-robin / load balancing modes

<p align="right">(<a href="#readme-top">back to top</a>)</p>



## Contributing

Contributions are welcome!

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

<p align="right">(<a href="#readme-top">back to top</a>)</p>


## License

Distributed under the Unlicense. See `LICENSE` for details.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- SHIELDS -->

[contributors-shield]: https://img.shields.io/github/contributors/othneildrew/Best-README-Template.svg?style=for-the-badge
[contributors-url]: https://github.com/othneildrew/Best-README-Template/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/othneildrew/Best-README-Template.svg?style=for-the-badge
[forks-url]: https://github.com/othneildrew/Best-README-Template/network/members
[license-shield]: https://img.shields.io/badge/License-RPL-blue.svg?style=for-the-badge
[license-url]: https://github.com/rubrionmc/gogate/blob/main/LICENSE.txt
[stars-shield]: https://img.shields.io/github/stars/rubrionmc/gogate?style=for-the-badge
[stars-url]: https://github.com/rubrionmc/gogate/stargazers
[issues-shield]: https://img.shields.io/github/issues/rubrionmc/gogate?style=for-the-badge
[issues-url]: https://github.com/rubrionmc/gogate/issues
