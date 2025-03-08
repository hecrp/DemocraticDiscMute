#!/bin/bash

# Basic Docker commands for DiscMuteBot
# For detailed documentation, see utils/docker_build.md

# Build the image
build() {
    echo "🏗️ Building Docker image..."
    docker build -t discmutebot .
}

# Run the container
run() {
    echo "🚀 Starting DiscMuteBot container..."
    docker run -d --name mute-bot \
        -v $(pwd)/config.json:/app/config.json \
        -v $(pwd)/logs:/app/logs \
        discmutebot
}

# Stop the container
stop() {
    echo "🛑 Stopping DiscMuteBot..."
    docker stop mute-bot
}

# Start an existing container
start() {
    echo "▶️ Starting DiscMuteBot..."
    docker start mute-bot
}

# Restart the container
restart() {
    echo "🔄 Restarting DiscMuteBot..."
    docker restart mute-bot
}

# View logs
logs() {
    echo "📋 Showing logs..."
    docker logs -f mute-bot
}

# Update bot (rebuild and restart)
update() {
    echo "🔄 Updating DiscMuteBot..."
    docker stop mute-bot
    docker rm mute-bot
    docker build -t discmutebot .
    docker run -d --name mute-bot \
        -v $(pwd)/config.json:/app/config.json \
        -v $(pwd)/logs:/app/logs \
        discmutebot
}

# Clean up (remove container and image)
cleanup() {
    echo "🧹 Cleaning up Docker resources..."
    docker stop mute-bot 2>/dev/null || true
    docker rm mute-bot 2>/dev/null || true
    docker rmi discmutebot 2>/dev/null || true
}

# Show help
help() {
    echo "DiscMuteBot Docker Management Script"
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  build   - Build Docker image"
    echo "  run     - Run container"
    echo "  stop    - Stop container"
    echo "  start   - Start existing container"
    echo "  restart - Restart container"
    echo "  logs    - View container logs"
    echo "  update  - Update bot (rebuild and restart)"
    echo "  cleanup - Remove container and image"
    echo "  help    - Show this help message"
}

# Main script logic
case "$1" in
    "build")
        build
        ;;
    "run")
        run
        ;;
    "stop")
        stop
        ;;
    "start")
        start
        ;;
    "restart")
        restart
        ;;
    "logs")
        logs
        ;;
    "update")
        update
        ;;
    "cleanup")
        cleanup
        ;;
    "help"|"")
        help
        ;;
    *)
        echo "Unknown command: $1"
        help
        exit 1
        ;;
esac 