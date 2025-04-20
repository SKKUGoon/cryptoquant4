# CryptoQuant Trading System

[![Go Version](https://img.shields.io/badge/go-1.23.1-blue.svg)](https://golang.org)
[![Docker](https://img.shields.io/badge/docker-%230db7ed.svg?style=flat&logo=docker&logoColor=white)](https://www.docker.com)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

A sophisticated cryptocurrency arbitrage and trading platform built with Go, focusing on price differences between Binance and Upbit exchanges.

## Overview

This system implements a high-frequency trading strategy that:
- Monitors and exploits price differences (premium) between Binance and Upbit exchanges
- Executes pair trades (long-short positions) when profitable opportunities are identified
- Supports multiple trading pairs with configurable parameters
- Implements real-time market data processing through WebSocket connections
- Features a unified account management system
- Provides comprehensive trade logging and analysis capabilities

## Project Structure

```
.
├── core/                  # Core trading engine and server implementation
│   ├── account/          # Unified account management
│   ├── trader/           # Exchange-specific trading implementations
│   │   ├── binance/      # Binance trading implementation
│   │   └── upbit/        # Upbit trading implementation
├── internal/             # Internal packages
│   ├── binance/          # Binance API clients
│   └── upbit/            # Upbit API clients
├── data/                 # Database and data management
│   ├── database/         # Database connections and models
│   └── timescale/        # Time series data management
├── signal/               # Signal processing and analysis
├── strategy/             # Trading strategies
├── proto/                # Protocol buffer definitions
├── gen/                  # Generated protocol buffer code
├── config/               # Configuration management
├── utils/                # Utility functions
└── main-*.go             # Entry points for different services
```

## Key Features

- **Unified Account Management**: Centralized account handling across multiple exchanges
- **Precision Trading**: Configurable precision settings for different exchanges
- **Real-time Monitoring**: WebSocket-based market data streaming
- **Trade Execution**: Automated order placement with safety margins
- **Data Analysis**: Comprehensive trade logging and premium tracking
- **Docker Support**: Development and production environments

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

# Trading Configuration
SAFE_MARGIN=0.9  # Safety margin for order execution
```

## Development Setup

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
- Database services
- Trading services
- Signal processing services

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

The project supports multiple build targets:

- `init`: Initializes the system and performs setup tasks
- `trader`: Runs the main trading services
- `signal`: Runs signal processing services
- `local`: Development environment with debugging tools

## Architecture

### Core Components

1. **Trading Server**
   - gRPC-based server implementation
   - Unified account management
   - Order execution and position management
   - Trade logging and analysis

2. **Exchange Integration**
   - Binance and Upbit API clients
   - WebSocket stream handlers
   - Order book management
   - Precision handling

3. **Data Management**
   - PostgreSQL for relational data
   - TimescaleDB for time series data
   - Redis for caching and real-time data

### Security Features

- Secure API key management
- Encrypted WebSocket connections
- Rate limiting and error handling
- Safe margin implementation for order execution

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
