#!/bin/bash

# Raidark Project Update Script
# This script updates the project with latest Raidark changes

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
RAIDARK_REPO="https://github.com/r0x16/Raidark.git"
TEMP_DIR="/tmp/raidark_update_$(date +%s)"
CURRENT_DIR="$(pwd)"

# Function to display usage
show_usage() {
    echo -e "${BLUE}Raidark Project Update${NC}"
    echo ""
    echo "Usage: $0 update [options]"
    echo ""
    echo "Options:"
    echo "  --dry-run    Show what would be updated without making changes"
    echo "  --force      Force update without confirmation"
    echo ""
    echo "This command will:"
    echo "  - Download latest Raidark repository"
    echo "  - Preserve your project configuration"
    echo "  - Update core files and dependencies"
    echo "  - Maintain your project name and customizations"
    echo ""
    echo "Example:"
    echo "  $0 update"
    echo "  $0 update --dry-run"
}

# Function to cleanup temporary directory
cleanup() {
    if [ -d "$TEMP_DIR" ]; then
        echo -e "${YELLOW}Cleaning up temporary files...${NC}"
        rm -rf "$TEMP_DIR"
    fi
}

# Set trap to cleanup on exit
trap cleanup EXIT

# Function to get current project name from go.mod
get_project_name() {
    if [ -f "go.mod" ]; then
        grep "^module" go.mod | awk '{print $2}' | sed 's|.*/||'
    else
        echo "unknown-project"
    fi
}

# Function to download latest Raidark
download_raidark() {
    echo -e "${BLUE}Downloading latest Raidark repository...${NC}"
    
    git clone "$RAIDARK_REPO" "$TEMP_DIR"
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}Successfully downloaded Raidark${NC}"
    else
        echo -e "${RED}Failed to download Raidark repository${NC}"
        exit 1
    fi
}

# Function to backup current files
backup_current_files() {
    local backup_dir="backup_$(date +%Y%m%d_%H%M%S)"
    
    echo -e "${BLUE}Creating backup of current files...${NC}"
    
    mkdir -p "$backup_dir"
    
    # Backup important files
    local files_to_backup=(
        "go.mod"
        "go.sum"
        ".env"
        ".env-example"
        "README.md"
        "main/"
        "shared/"
    )
    
    for file in "${files_to_backup[@]}"; do
        if [ -e "$file" ]; then
            cp -r "$file" "$backup_dir/"
            echo -e "${YELLOW}Backed up: $file${NC}"
        fi
    done
    
    echo -e "${GREEN}Backup created in: $backup_dir${NC}"
}

# Function to update core files
update_core_files() {
    local project_name="$1"
    local dry_run="$2"
    
    echo -e "${BLUE}Updating core files...${NC}"
    
    # Files to update from Raidark
    local files_to_update=(
        "main/"
        "shared/"
        "Dockerfile"
        "docker-compose.yml"
        ".gitignore"
    )
    
    for file in "${files_to_update[@]}"; do
        if [ -e "$TEMP_DIR/$file" ]; then
            if [ "$dry_run" = "true" ]; then
                echo -e "${YELLOW}Would update: $file${NC}"
            else
                echo -e "${YELLOW}Updating: $file${NC}"
                cp -r "$TEMP_DIR/$file" "$CURRENT_DIR/"
            fi
        fi
    done
}

# Function to update go.mod dependencies
update_dependencies() {
    local project_name="$1"
    local dry_run="$2"
    
    echo -e "${BLUE}Updating Go dependencies...${NC}"
    
    if [ "$dry_run" = "true" ]; then
        echo -e "${YELLOW}Would update go.mod dependencies${NC}"
        return
    fi
    
    # Preserve current module name
    local current_module=$(grep "^module" go.mod | awk '{print $2}')
    
    # Copy new go.mod but preserve module name
    if [ -f "$TEMP_DIR/go.mod" ]; then
        cp "$TEMP_DIR/go.mod" go.mod
        sed -i "1s|^module.*|module $current_module|" go.mod
        echo -e "${GREEN}Updated go.mod with preserved module name${NC}"
    fi
    
    # Update dependencies
    go mod tidy
    echo -e "${GREEN}Updated Go dependencies${NC}"
}

# Function to update import statements
update_imports() {
    local project_name="$1"
    local dry_run="$2"
    
    echo -e "${BLUE}Updating import statements...${NC}"
    
    if [ "$dry_run" = "true" ]; then
        echo -e "${YELLOW}Would update import statements in Go files${NC}"
        return
    fi
    
    local current_module=$(grep "^module" go.mod | awk '{print $2}')
    
    # Find all Go files and update imports
    find . -name "*.go" -type f | while read -r file; do
        if grep -q "github.com/r0x16/Raidark" "$file"; then
            echo -e "${YELLOW}Updating imports in $file...${NC}"
            sed -i "s|github.com/r0x16/Raidark|$current_module|g" "$file"
        fi
    done
    
    echo -e "${GREEN}Updated import statements${NC}"
}

# Function to show update summary
show_update_summary() {
    echo ""
    echo -e "${GREEN}Update Summary:${NC}"
    echo "  ✓ Core files updated"
    echo "  ✓ Dependencies updated"
    echo "  ✓ Import statements updated"
    echo "  ✓ Project configuration preserved"
    echo ""
    echo -e "${BLUE}Next steps:${NC}"
    echo "  1. Test your application: go run main/main.go"
    echo "  2. Review changes and commit if everything works"
    echo "  3. Update your .env file if needed"
}

# Function to check if we're in a Raidark project
check_raidark_project() {
    if [ ! -f "go.mod" ]; then
        echo -e "${RED}Error: No go.mod found. Are you in a Go project directory?${NC}"
        exit 1
    fi
    
    if [ ! -d "main" ] || [ ! -d "shared" ]; then
        echo -e "${RED}Error: This doesn't appear to be a Raidark project${NC}"
        echo -e "${YELLOW}Expected directories 'main' and 'shared' not found${NC}"
        exit 1
    fi
}

# Main execution
main() {
    local dry_run=false
    local force=false
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --dry-run)
                dry_run=true
                shift
                ;;
            --force)
                force=true
                shift
                ;;
            update)
                shift
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            *)
                echo -e "${RED}Unknown option: $1${NC}"
                show_usage
                exit 1
                ;;
        esac
    done
    
    check_raidark_project
    
    local project_name=$(get_project_name)
    
    echo -e "${GREEN}Updating Raidark project: $project_name${NC}"
    echo ""
    
    if [ "$dry_run" = "true" ]; then
        echo -e "${YELLOW}DRY RUN MODE - No changes will be made${NC}"
        echo ""
    elif [ "$force" = "false" ]; then
        echo -e "${YELLOW}Warning: This will update core files with latest Raidark changes.${NC}"
        echo -e "${YELLOW}Your project-specific code will be preserved, but core files will be updated.${NC}"
        read -p "Are you sure you want to continue? (y/N): " -n 1 -r
        echo
        
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo -e "${RED}Operation cancelled${NC}"
            exit 1
        fi
    fi
    
    download_raidark
    
    if [ "$dry_run" = "false" ]; then
        backup_current_files
    fi
    
    update_core_files "$project_name" "$dry_run"
    update_dependencies "$project_name" "$dry_run"
    update_imports "$project_name" "$dry_run"
    
    if [ "$dry_run" = "false" ]; then
        show_update_summary
    else
        echo ""
        echo -e "${BLUE}Dry run completed. Use 'update' without --dry-run to apply changes.${NC}"
    fi
}

# Execute main function with all arguments
main "$@" 