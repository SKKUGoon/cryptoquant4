# CryptoQuant Trading System

[![Go Version](https://img.shields.io/badge/go-1.23.1-blue.svg)](https://golang.org)
[![Docker](https://img.shields.io/badge/docker-%230db7ed.svg?style=flat&logo=docker&logoColor=white)](https://www.docker.com)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

A cryptocurrency quantitative analysis and trading platform built with Go.

## Overview

This system implements a trading strategy that:
- Monitors price differences between Binance and Upbit exchanges
- Executes trades when profitable opportunities are identified
- Supports multiple trading pairs (XRP, COW, etc.)
- Uses Redis for caching and real-time data
- Implements WebSocket connections for real-time market data
- Supports both development and production environments

## Project Structure

```
.
├── engine/                 # Core trading engine
│   ├── trade/             # Exchange-specific trading implementations
│   │   ├── binance/       # Binance trading implementation
│   │   └── upbit/         # Upbit trading implementation
│   └── strategy/          # Trading strategies
├── internal/              # Internal packages
│   ├── binance/           # Binance API clients
│   └── upbit/             # Upbit API clients
├── streams/               # WebSocket stream handlers
│   ├── binance/           # Binance WebSocket streams
│   └── upbit/             # Upbit WebSocket streams
└── main-*.go              # Entry points for different build targets
```

## Prerequisites

- Go 1.23.1 or later
- Docker and Docker Compose
- Redis
- PostgreSQL
- TimescaleDB
- API keys for Binance and Upbit

## Environment Variables

Create a `.env` file with the following variables:

```bash
# Database Configuration
USERNAME=your_db_username
PASSWORD=your_db_password
PG_HOST=your_postgres_host
PG_PORT=your_postgres_port
PG_NAME=your_db_name
TS_HOST=your_timescaledb_host
TS_PORT=your_timescaledb_port
TS_NAME=your_ts_db_name

# Exchange API Keys
BINANCE_API_KEY=your_binance_api_key
BINANCE_SECRET_KEY=your_binance_secret_key
UPBIT_API_KEY=your_upbit_api_key
UPBIT_SECRET_KEY=your_upbit_secret_key

# Trading Pairs
BINANCE_SYMBOL=XRPUSDT
UPBIT_SYMBOL=KRW-XRP
ANCHOR_SYMBOL=KRW-USDT
```

## Development Setup

Prerequisites are:
- Postgres 
- Timescale DB

1. Create the Docker network:
```bash
docker network create crypto-dev
```

2. Start the development environment:
```bash
docker compose -f docker-compose.dev.yaml up --build
```

This will start:
- Redis for caching
- Initialization service
- Trading services for each pair (XRP, COW, etc.)

## Production Setup

1. Create the Docker network:
```bash
docker network create crypto-bridge
```

2. Start the production environment:
```bash
docker-compose up --build
```

## Build Targets

The project supports different build targets:

- `init`: Initializes Redis and performs setup tasks
- `server`: Runs the trading services
- Default: Development environment with debugging tools

## Monitoring

- Development: Access pprof at `localhost:6060`
- Logs are stored in `./log/{engine_name}/`
- Redis metrics available on port 6379

## Architecture

### Trading Engine
- Implements high-frequency trading strategies
- Monitors multiple trading pairs simultaneously
- Uses WebSocket connections for real-time data
- Implements order book management
- Handles trade execution and position management

### Data Flow
1. WebSocket streams receive real-time market data
2. Trading engine processes data and identifies opportunities
3. Orders are executed when profitable conditions are met
4. Results are logged to TimescaleDB for analysis

### Security
- API keys are managed through environment variables
- Secure WebSocket connections
- Rate limiting and error handling
- Proper cleanup of sensitive data

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
