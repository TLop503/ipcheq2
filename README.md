# ipcheq2

Aggregate data from AbuseIPDB, VPNID, and VirusTotal to investigate IPs!

![ipcheq2 homescreen with search bar](image.png)

## Features
- Search results history!
- Full IPv4 Support, with v6 coverage coming soon!
- Extendable data-- bring your own sources to supplement ours!
- Concise, distraction-free results without any marketing!
- Linux and Windows both supported!
- Choose between serving a Web UI, Headless API, or both!
  - Use `-h` for details on flags

### IP Data Sources:
- AbuseIPDB
- VirusTotal (Optional)
- iCloud Private Relays
- Cyberghost
- Express VPN
- Mullvad VPN
- Nord VPN
- PIA
- Proton VPN
- Surfshark VPN
- Tor Exit Nodes
- Tunnelbear VPN

Want to see another source here? Open a PR with a file or an issue with a link!

## Quick Start with Docker

### Using Docker Run
```bash
# supply api keys
docker run -p 8080:8080 -e ABIPDBKEY=your_api_key_here -e VTKEY=your_api_key_here ghcr.io/tlop503/ipcheq2:latest <optional --mode webui|api|headless>
```

### Using Docker Compose
1. Create a `docker-compose.yml` file containing:
```yaml
version: '3.8'
services:
  ipcheq2:
    image: ghcr.io/tlop503/ipcheq2:latest
    command: ["--mode", "webui|api|headless"] # optional
    ports:
      - "8080:8080"
    environment:
      - ABIPDBKEY=your_api_key_here
      - VTKEY=your_api_key_here # Optional
    restart: unless-stopped
```

2. Run the application:
```bash
docker-compose up -d
```

## Run locally
1. Download exe or elf from the latest [release](https://github.com/tlop503/ipcheq2/releases/latest).
	   VPN source lists and config are bundled into the binary, so no extra data files are required.
```
└── ipcheq2 or ipcheq2.exe
```
2. Create a .env file with AbuseIPDB and VirusTotal API Keys (see `.env.example`) in the same directory, or set an enviornment variable.
3. Optionally update iCloud prefixes before building your own binary.
  1. Use the Go updater from the repository root:
  ```bash
  go run ./cmd/update-icloud -data-dir ./internal/data
  ```
4. Run the executable! ipcheq2 will serve on localhost:8080.

## Development Setup

### Prerequisites
- Go 1.23+
- AbuseIPDB API key
- VirusTotal API key (optional)

### Local Development
1. Clone the repository:
```bash
git clone https://github.com/tlop503/ipcheq2.git
cd ipcheq2
```

2. Create a `.env` file:
```bash
cp .env.example .env
# Edit .env and add your ABIPDBKEY, VTKEY
# Alternatively, enviornment variables can be used.
```

3. Run the application:
```bash
go run main.go
# or
go build . # inside project
./ipcheq2
```

4. Open your browser to `http://localhost:8080`

#### Updating iCloud Private Relay prefixes

The updater can be run with `go run ./cmd/update-icloud -data-dir ./internal/data`.

## Deployment

### GitHub Container Registry
This project is automatically built and published to GitHub Container Registry via GitHub Actions.

### Manual Build
```bash
docker build -t ipcheq2 ./dockerfiles
docker run -p 8080:8080 -e ABIPDBKEY=your_api_key_here VTKEY=your_api_key_here ipcheq2
```

Data Sources:  
vpn_by_asn retrieved from https://github.com/X4BNet/lists_vpn. Copyright (c) 2024 X4B (Mathew Heard).
Supplemented with https://github.com/NazgulCoder/IPLists?tab=readme-ov-file. Copyright (c) 2025 Nazgul.
