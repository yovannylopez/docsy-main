#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Load environment variables
source ../.env

# Build the connection URL
DB_URL="postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=$DB_SSLMODE"

# Function to show the usage of the script
show_usage() {
    echo "Usage: $0 <command> [version]"
    echo "Available commands:"
    echo "  create <name>  - Create a new migration"
    echo "  up [n]          - Apply all migrations or n migrations"
    echo "  down [n]        - Revert all migrations or n migrations"
    echo "  force <version> - Force the database version"
    echo "  version        - Show the current version of the database"
    echo "  status         - Show the status of the migrations"
}

# Verify that a command was provided
if [ $# -lt 1 ]; then
    show_usage
    exit 1
fi

COMMAND=$1
shift

case $COMMAND in
    "create")
        if [ $# -lt 1 ]; then
            echo "Error: You must provide a name for the migration"
            exit 1
        fi
        migrate create -ext sql -dir ../migrations/core -seq "${1}"
        ;;
    "up")
        if [ $# -eq 1 ]; then
            migrate -database "${DB_URL}" -path ../migrations/core up $1
        else
            migrate -database "${DB_URL}" -path ../migrations/core up
        fi
        ;;
    "down")
        if [ $# -eq 1 ]; then
            migrate -database "${DB_URL}" -path ../migrations/core down $1
        else
            migrate -database "${DB_URL}" -path ../migrations/core down
        fi
        ;;
    "force")
        if [ $# -ne 1 ]; then
            echo "Error: You must provide a version"
            exit 1
        fi
        migrate -database "${DB_URL}" -path ../migrations/core force $1
        ;;
    "version")
        migrate -database "${DB_URL}" -path ../migrations/core version
        ;;
    "status")
        migrate -database "${DB_URL}" -path ../migrations/core status
        ;;
    *)
        echo "Unknown command: $COMMAND"
        show_usage
        exit 1
        ;;
esac
