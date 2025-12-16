# ipcheq2

Aggregate data from AbuseIPDB and VPNID to investigate IPs!

![ipcheq2 homescreen with search bar](image.png)

## Features
- Search results history!
- Full IPv4 Support, with v6 coverage coming soon!
- Extendable data-- bring your own sources to supplement ours!
- Concise, distraction-free results without any marketing!
- Linux and Windows both supported!

### IP Data Sources:
- AbuseIPDB
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
# supply an abuseipdb api key
docker run -p 8080:8080 -e ABIPDBKEY=your_api_key_here ghcr.io/tlop503/ipcheq2:latest
```

### Using Docker Compose
1. Create a `docker-compose.yml` file:
```yaml
version: '3.8'
services:
  ipcheq2:
    image: ghcr.io/tlop503/ipcheq2:latest
    ports:
      - "8080:8080"
    environment:
      - ABIPDBKEY=your_api_key_here
    restart: unless-stopped
```

2. Run the application:
```bash
docker-compose up -d
```

## Run locally
1. Download exe or elf from the latest [release](https://github.com/tlop503/ipcheq2/releases/latest). Download and extract the ip data as well and arrange to match tree:
   	Alternatively clone the repo and build from source to skip arranging files (see "Local Development" below)
```
├── ipcheq2 or ipcheq2.exe
└── data
        ├── source1.txt
        ├── ...
        ├── update_icloud_relays.py
        └── upstream-icloud-list.hash
└── vpnid_config.txt
```
2. Create a .env file with an AbuseIPDB API Key (see `.env.example`) in the same directory, or set an enviornment variable.
3. Update the icloud prefixes if desired with the bundled Python script.
	1. Note- this must be ran from wihtin the `data/` directory!
4. Optionally add IP lists to data/ and update the config file to match
5. Run the executable! ipcheq2 will serve on localhost:8080.

## Development Setup

### Prerequisites
- Go 1.23+
- AbuseIPDB API key

### Local Development
1. Clone the repository:
```bash
git clone https://github.com/tlop503/ipcheq2.git
cd ipcheq2
```

2. Create a `.env` file:
```bash
cp .env.example .env
# Edit .env and add your ABIPDBKEY
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

[A script is provided](data/update_icloud_relays.py) to update the iCloud Private Relay prefixes within the repo.
Before running it, make sure that you're inside the `data/` directory.

## Deployment

### GitHub Container Registry
This project is automatically built and published to GitHub Container Registry via GitHub Actions.

### Manual Build
```bash
docker build -t ipcheq2 .
docker run -p 8080:8080 -e ABIPDBKEY=your_api_key_here ipcheq2
```
