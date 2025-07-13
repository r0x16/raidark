#!/bin/bash

# Raidark Project Creation Script
# This script cleans the environment and converts it to a new project

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to display usage
show_usage() {
    echo -e "${BLUE}Raidark Project Creation${NC}"
    echo ""
    echo "Usage: $0 create <project_name>"
    echo ""
    echo "This command will:"
    echo "  - Clean the current environment"
    echo "  - Remove Raidark-specific files"
    echo "  - Rename the project to your specified name"
    echo "  - Update go.mod with new module name"
    echo "  - Initialize as a new project"
    echo ""
    echo "Example:"
    echo "  $0 create MyAwesomeProject"
}

# Function to validate project name
validate_project_name() {
    local project_name="$1"
    
    if [ -z "$project_name" ]; then
        echo -e "${RED}Error: Project name is required${NC}"
        show_usage
        exit 1
    fi
    
    # Check if project name contains only valid characters
    if [[ ! "$project_name" =~ ^[a-zA-Z0-9_-]+$ ]]; then
        echo -e "${RED}Error: Project name can only contain letters, numbers, hyphens, and underscores${NC}"
        exit 1
    fi
}

# Function to clean Raidark-specific files
clean_raidark_files() {
    echo -e "${BLUE}Cleaning Raidark-specific files...${NC}"
    
    # Remove git history
    if [ -d ".git" ]; then
        echo -e "${YELLOW}Removing git history...${NC}"
        rm -rf .git
    fi
    
    # Remove Raidark-specific files
    local files_to_remove=(
        "README.md"
        "LICENSE"
        "raidark.go"
        "scripts"
        "raidark.sh"
    )
    
    for file in "${files_to_remove[@]}"; do
        if [ -e "$file" ]; then
            echo -e "${YELLOW}Removing $file...${NC}"
            rm -rf "$file"
        fi
    done
}

# Function to update go.mod with new module name
update_go_mod() {
    local project_name="$1"
    local new_module_name="github.com/$(whoami)/$project_name"
    
    echo -e "${BLUE}Updating go.mod with new module name: $new_module_name${NC}"
    
    if [ -f "go.mod" ]; then
        # Replace the module name in go.mod
        sed -i "1s|^module.*|module $new_module_name|" go.mod
        echo -e "${GREEN}Updated go.mod successfully${NC}"
    else
        echo -e "${YELLOW}go.mod not found, creating new one...${NC}"
        go mod init "$new_module_name"
    fi
}

# Function to update import statements in Go files
update_imports() {
    local project_name="$1"
    local new_module_name="github.com/$(whoami)/$project_name"
    
    echo -e "${BLUE}Updating import statements in Go files...${NC}"
    
    # Find all Go files and update imports
    find . -name "*.go" -type f | while read -r file; do
        if grep -q "github.com/r0x16/Raidark" "$file"; then
            echo -e "${YELLOW}Updating imports in $file...${NC}"
            sed -i "s|github.com/r0x16/Raidark|$new_module_name|g" "$file"
        fi
    done
}

# Function to create basic project structure
create_project_structure() {
    local project_name="$1"
    
    echo -e "${BLUE}Creating basic project structure...${NC}"
    
    # Create basic README
    cat > README.md << EOF
# $project_name

A project built with Raidark architecture.

## Getting Started

This project follows hexagonal architecture principles with clean separation of concerns.

### Structure

- \`main/\` - Application entry point and configuration
- \`shared/\` - Shared utilities and common code

### Running the Project

\`\`\`bash
go run main/main.go
\`\`\`

### Environment Variables

Copy \`.env-example\` to \`.env\` and configure your environment variables.

## Development

This project was bootstrapped with [Raidark](https://github.com/r0x16/Raidark).
EOF
    
    echo -e "${GREEN}Created README.md${NC}"
}

# Function to initialize new git repository
init_git_repo() {
    local project_name="$1"
    
    echo -e "${BLUE}Initializing new git repository...${NC}"
    
    git init
    git add .
    git commit -m "feat: Initial commit for $project_name project

Created new project based on Raidark architecture.

Changes:
    - Initialized project structure
    - Updated module name
    - Created basic README
    - Configured environment"
    
    echo -e "${GREEN}Git repository initialized successfully${NC}"
}

# Main execution
main() {
    local project_name="$2"
    
    # Skip the 'create' argument
    if [ "$1" = "create" ]; then
        shift
        project_name="$1"
    fi
    
    validate_project_name "$project_name"
    
    echo -e "${GREEN}Creating new Raidark project: $project_name${NC}"
    echo ""
    
    # Confirm with user
    echo -e "${YELLOW}Warning: This will remove git history and Raidark-specific files.${NC}"
    read -p "Are you sure you want to continue? (y/N): " -n 1 -r
    echo
    
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${RED}Operation cancelled${NC}"
        exit 1
    fi
    
    clean_raidark_files
    update_go_mod "$project_name"
    update_imports "$project_name"
    create_project_structure "$project_name"
    init_git_repo "$project_name"
    
    echo ""
    echo -e "${GREEN}✓ Project '$project_name' created successfully!${NC}"
    echo -e "${BLUE}You can now start developing your application.${NC}"
    echo ""
    echo "Next steps:"
    echo "  1. Review and update .env file"
    echo "  2. Start coding your application logic"
    echo "  3. Run: go run main/main.go"
}

# Execute main function with all arguments
main "$@" 