# IGDB Database Collector

A service that collects and stores video game data from IGDB (Internet Game Database) into MongoDB. It provides both bulk data collection and webhook-based real-time updates.

## Features

- Bulk collection of IGDB data
- Real-time updates via webhooks
- MongoDB storage
- Support for all IGDB endpoints

## Prerequisites

- Go 1.24.1 or higher
- MongoDB
- IGDB API credentials (Client ID and Secret from Twitch)

## Configuration

Create a `config.json` file in the root directory with the following structure:

```json
{
  "address": "localhost:8080",
  "database": {
    "host": "localhost",
    "port": 27017,
    "user": "username",
    "password": "password",
    "database": "igdb"
  },
  "twitch": {
    "client_id": "your_client_id",
    "client_secret": "your_client_secret"
  },
  "webhook_secret": "your_webhook_secret",
  "external_url": "https://your-webhook-url.com"
}
```

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/igdb-database.git

# Install dependencies
go mod download
```

## Usage

### Running the Service

```bash
go run main.go
```

The service will:

1. Connect to MongoDB
2. Initialize IGDB client
3. Fetch initial data if collections are empty
4. Start webhook server for real-time updates

## Dependencies

- [go-igdb](https://github.com/bestnite/go-igdb) - IGDB API client
- [mongo-driver](https://github.com/mongodb/mongo-go-driver) - MongoDB driver for Go

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
