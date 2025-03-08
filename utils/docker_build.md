# 🐳 Docker Guide for DiscMuteBot

This guide explains how to build, run, and manage your Discord bot using Docker. For basic setup, refer to the Docker section in the main README.md file. This document provides extended information and advanced usage details.

## 📋 Prerequisites

- Docker installed on your system
- `config.json` file with your Discord bot token
- Basic understanding of terminal/command line

## 🛠️ Helper Script

This directory includes a helper script (`docker_build.sh`) that simplifies Docker operations. To use it:

1. Make the script executable:
   ```bash
   chmod +x utils/docker_build.sh
   ```

2. View available commands:
   ```bash
   ./utils/docker_build.sh help
   ```

### Available Commands

- `build` - Build Docker image
- `run` - Run container
- `stop` - Stop container
- `start` - Start existing container
- `restart` - Restart container
- `logs` - View container logs
- `update` - Update bot (rebuild and restart)
- `cleanup` - Remove container and image

### Example Usage

```bash
# Build the image
./utils/docker_build.sh build

# Start the bot
./utils/docker_build.sh run

# View logs
./utils/docker_build.sh logs

# Update the bot after code changes
./utils/docker_build.sh update

# Clean up everything
./utils/docker_build.sh cleanup
```

## Manual Docker operations:

## 🏗️ Building the Bot

Build the Docker image:
```bash
docker build -t discmutebot .
```

This command:
- Creates a new Docker image named 'discmutebot'
- Uses the multi-stage build process defined in Dockerfile
- Optimizes the final image size

## 🚀 Running the Bot

### First Time Setup
```bash
docker run -d --name mute-bot \
  -v $(pwd)/config.json:/app/config.json \
  -v $(pwd)/logs:/app/logs \
  discmutebot
```

Parameters explained:
- `-d`: Run in detached mode (background)
- `--name mute-bot`: Assign a name to the container
- `-v $(pwd)/config.json:/app/config.json`: Mount configuration file
- `-v $(pwd)/logs:/app/logs`: Mount logs directory

## 📊 Monitoring

### View Container Logs
```bash
docker logs mute-bot
```

### Check Container Status
```bash
docker ps
```

## 🎮 Container Management

### Basic Controls
```bash
# Stop the bot
docker stop mute-bot

# Start the bot (after stopping)
docker start mute-bot

# Restart the bot
docker restart mute-bot
```

### Cleanup
```bash
# Remove the container (must be stopped first)
docker rm mute-bot

# Remove the image
docker rmi discmutebot
```

## 🔄 Updating the Bot

When you make code changes, follow these steps:

1. **Rebuild the image**:
```bash
docker build -t discmutebot .
```

2. **Remove old container**:
```bash
docker stop mute-bot
docker rm mute-bot
```

3. **Create new container**:
```bash
docker run -d --name mute-bot \
  -v $(pwd)/config.json:/app/config.json \
  -v $(pwd)/logs:/app/logs \
  discmutebot
```

## 🔍 Troubleshooting

If the bot isn't working:

1. Check if the container is running:
   ```bash
   docker ps | grep mute-bot
   ```

2. View recent logs:
   ```bash
   docker logs --tail 50 mute-bot
   ```

3. Verify your config.json is mounted correctly:
   ```bash
   docker exec mute-bot ls -l /app/config.json
   ```

4. Restart the container:
   ```bash
   docker restart mute-bot
   ```

## 🛠️ Advanced Usage

### Interactive Shell
```bash
docker exec -it mute-bot /bin/sh
```

### Real-time Log Following
```bash
docker logs -f mute-bot
```

### Container Resource Usage
```bash
docker stats mute-bot
```