#!/bin/bash

# Raidark Project Management Script
# This script acts as a dispatcher for various Raidark project operations

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCRIPTS_DIR="$SCRIPT_DIR/scripts"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to display usage
show_usage() {
    echo -e "${BLUE}Raidark Project Management Script${NC}"
    echo ""
    echo "Usage: $0 <command> [options]"
    echo ""
    echo "Commands:"
    echo -e "  ${GREEN}create${NC}    Create a new Raidark project (clean environment and convert to new project)"
    echo -e "  ${GREEN}update${NC}    Update project with latest Raidark changes"
    echo -e "  ${GREEN}codex${NC}     Generate AGENTS.md file for Codex tool"
    echo -e "  ${GREEN}help${NC}      Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 create MyProject"
    echo "  $0 update"
    echo "  $0 codex"
}

# Function to check if scripts directory exists
check_scripts_dir() {
    if [ ! -d "$SCRIPTS_DIR" ]; then
        echo -e "${RED}Error: Scripts directory not found at $SCRIPTS_DIR${NC}"
        exit 1
    fi
}

# Function to execute script with error handling
execute_script() {
    local script_name="$1"
    local script_path="$SCRIPTS_DIR/$script_name"
    
    if [ ! -f "$script_path" ]; then
        echo -e "${RED}Error: Script $script_name not found at $script_path${NC}"
        exit 1
    fi
    
    if [ ! -x "$script_path" ]; then
        echo -e "${YELLOW}Making script executable: $script_name${NC}"
        chmod +x "$script_path"
    fi
    
    echo -e "${BLUE}Executing: $script_name${NC}"
    shift
    "$script_path" "$@"
}

# Main command processing
case "${1:-help}" in
    "create")
        check_scripts_dir
        execute_script "create-project.sh" "$@"
        ;;
    "update")
        check_scripts_dir
        execute_script "update-project.sh" "$@"
        ;;
    "codex")
        check_scripts_dir
        execute_script "create-codex.sh" "$@"
        ;;
    "help"|"-h"|"--help")
        show_usage
        ;;
    *)
        echo -e "${RED}Error: Unknown command '$1'${NC}"
        echo ""
        show_usage
        exit 1
        ;;
esac 