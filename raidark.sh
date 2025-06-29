#!/bin/bash

# Color definitions for rich console UI
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# ASCII Art function
print_logo() {
        echo -e "${CYAN}"
    cat << "EOF"
    ╔═══════════════════════════════════════════════════════════╗
    ║                                                           ║
    ║    ██████╗  █████╗ ██╗██████╗  █████╗ ██████╗ ██╗  ██╗    ║
    ║    ██╔══██╗██╔══██╗██║██╔══██╗██╔══██╗██╔══██╗██║ ██╔╝    ║
    ║    ██████╔╝███████║██║██║  ██║███████║██████╔╝█████╔╝     ║
    ║    ██╔══██╗██╔══██║██║██║  ██║██╔══██║██╔══██╗██╔═██╗     ║
    ║    ██║  ██║██║  ██║██║██████╔╝██║  ██║██║  ██║██║  ██╗    ║
    ║    ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝    ║
    ║                                                           ║
    ║              🚀 PROJECT AUTOMATION TOOLKIT 🚀             ║
    ║                                                           ║
    ╚═══════════════════════════════════════════════════════════╝
EOF
    echo -e "${NC}"
}

# Function to print section headers
print_section() {
    echo -e "\n${PURPLE}${BOLD}╔═══════════════════════════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${PURPLE}${BOLD}║  $1${NC}"
    echo -e "${PURPLE}${BOLD}╚═══════════════════════════════════════════════════════════════════════════════════════╝${NC}\n"
}

# Function to print success messages
print_success() {
    echo -e "${GREEN}${BOLD}✅ $1${NC}"
}

# Function to print error messages
print_error() {
    echo -e "${RED}${BOLD}❌ ERROR: $1${NC}"
}

# Function to print info messages
print_info() {
    echo -e "${CYAN}${BOLD}ℹ️  $1${NC}"
}

# Function to print warning messages
print_warning() {
    echo -e "${YELLOW}${BOLD}⚠️  $1${NC}"
}

# Function to get user input with styled prompt
get_user_input() {
    local prompt="$1"
    local variable_name="$2"
    echo -e "${WHITE}${BOLD}📝 $prompt${NC}"
    echo -ne "${YELLOW}➤ ${NC}"
    read -r $variable_name
}

# Function to get confirmation from user
get_confirmation() {
    local prompt="$1"
    echo -e "${WHITE}${BOLD}❓ $prompt (y/N)${NC}"
    echo -ne "${YELLOW}➤ ${NC}"
    read -r confirmation
    case $confirmation in
        [yY][eE][sS]|[yY])
            return 0
            ;;
        *)
            return 1
            ;;
    esac
}

# Function to show progress
show_progress() {
    local message="$1"
    echo -e "${BLUE}${BOLD}🔄 $message...${NC}"
}

# Function to validate module name format
validate_module_name() {
    local module_name="$1"
    if [[ ! $module_name =~ ^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}/[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+$ ]]; then
        return 1
    fi
    return 0
}

# Function to clean and setup new project
new_project() {
    print_section "🏗️  CREATE NEW PROJECT FROM RAIDARK BASE"
    
    # Check if we're in a Go project
    if [[ ! -f "go.mod" ]]; then
        print_error "This doesn't appear to be a Go project (no go.mod found)"
        exit 1
    fi
    
    echo -e "${CYAN}${BOLD}Welcome to Raidark Project Generator!${NC}"
    echo -e "${WHITE}This tool will help you create a new Go project using Raidark as a base template.${NC}"
    echo -e "${WHITE}Your new project will be completely independent from the original Raidark.${NC}\n"
    
    print_warning "This will clean the current directory and setup your new project!"
    if ! get_confirmation "Ready to create your new project?"; then
        print_info "Project creation cancelled"
        exit 0
    fi
    
    # Step 1: Get project name
    print_section "📋 PROJECT CONFIGURATION"
    echo -e "${CYAN}Let's configure your new project:${NC}\n"
    
    # Get current directory name as default suggestion
    current_dir=$(basename "$(pwd)")
    
    while true; do
        echo -e "${WHITE}${BOLD}📝 What's your new project name? ${CYAN}(default: ${current_dir})${NC}"
        echo -ne "${YELLOW}➤ ${NC}"
        read -r project_name
        
        # Use default if empty
        if [[ -z "$project_name" ]]; then
            project_name="$current_dir"
        fi
        
        # Convert to lowercase for consistency
        project_name=$(echo "$project_name" | tr '[:upper:]' '[:lower:]')
        break
    done
    
    # Step 2: Get module name
    echo -e "\n${CYAN}${BOLD}Module Name Examples:${NC}"
    echo -e "${WHITE}  • github.com/username/repository${NC}"
    echo -e "${WHITE}  • gitlab.com/username/repository${NC}"
    echo -e "${WHITE}  • bitbucket.org/username/repository${NC}\n"
    
    while true; do
        get_user_input "Enter your Go module name (repository URL):" new_module
        if [[ -z "$new_module" ]]; then
            print_error "Module name cannot be empty"
            continue
        fi
        if ! validate_module_name "$new_module"; then
            print_error "Invalid module name format. Please use: domain.com/user/repo"
            continue
        fi
        break
    done
    
    # Get current module name
    current_module=$(grep "^module " go.mod | awk '{print $2}')
    
    # Show summary
    print_section "📊 PROJECT SUMMARY"
    echo -e "${WHITE}${BOLD}Creating new project:${NC}"
    echo -e "${CYAN}  • Project name: ${BOLD}$project_name${NC}"
    echo -e "${CYAN}  • Module: ${BOLD}$new_module${NC}"
    echo -e "${CYAN}  • Based on: ${BOLD}Raidark template${NC}\n"
    
    if ! get_confirmation "Create project with these settings?"; then
        print_info "Project creation cancelled"
        exit 0
    fi
    
    print_section "🚀 CREATING YOUR PROJECT"
    
    show_progress "Setting up clean workspace"
    if [[ -d ".git" ]]; then
        rm -rf .git
    fi
    
    show_progress "Configuring Go module"
    sed -i "s|^module .*|module $new_module|" go.mod
    find . -name "*.go" -type f -exec sed -i "s|$current_module|$new_module|g" {} +
    
    show_progress "Customizing project files"
    
    # Get list of files to update (excluding binary files and git files)
    files_to_update=$(find . -type f \( -name "*.go" -o -name "*.md" -o -name "*.yml" -o -name "*.yaml" -o -name "*.json" -o -name "*.txt" -o -name "*.sh" -o -name "*.env*" -o -name "Dockerfile*" \) ! -path "./.git/*" ! -name "raidark.sh")
    
    total_replacements=0
    while IFS= read -r file; do
        if [[ -f "$file" ]]; then
            # Count replacements made - clean the output
            before_count=$(grep -ic "raidark" "$file" 2>/dev/null || echo "0")
            before_count=$(echo "$before_count" | tr -d '\n\r' | head -1)
            
            # Perform case-insensitive replacement
            sed -i 's/[Rr][Aa][Ii][Dd][Aa][Rr][Kk]/'$project_name'/g' "$file"
            
            if [[ "$before_count" =~ ^[0-9]+$ ]] && [[ $before_count -gt 0 ]]; then
                total_replacements=$((total_replacements + before_count))
            fi
        fi
    done <<< "$files_to_update"
    
    show_progress "Finalizing project setup"
    go mod tidy 2>/dev/null || true
    
    show_progress "Initializing Git repository"
    git init -b main 2>/dev/null || git init 2>/dev/null
    # Ensure we're on main branch (for older git versions)
    git checkout -b main 2>/dev/null || git branch -M main 2>/dev/null || true
    git add .
    
    # Create semantic initial commit
    git commit -m "feat(project): Initialize new $project_name project from Raidark template

Initialize new project based on Raidark template structure.
This commit establishes the base architecture and configuration for $project_name.

Changes:
    - Set up Go module as $new_module
    - Configure project structure from Raidark template
    - Initialize base project files and dependencies
    - Set up development environment configuration" 2>/dev/null || true
    
    # Final summary
    print_section "🎉 PROJECT CREATED SUCCESSFULLY!"
    echo -e "${GREEN}${BOLD}Welcome to your new $project_name project!${NC}\n"
    
    echo -e "${WHITE}${BOLD}Your project is ready:${NC}"
    echo -e "${GREEN}  ✅ Clean workspace prepared${NC}"
    echo -e "${GREEN}  ✅ Go module configured as: ${BOLD}$new_module${NC}"
    echo -e "${GREEN}  ✅ Project customized with name: ${BOLD}$project_name${NC}"
    echo -e "${GREEN}  ✅ Git repository initialized on ${BOLD}main${NC} branch"
    echo -e "${GREEN}  ✅ Initial commit created with semantic format${NC}\n"
    
    echo -e "${CYAN}${BOLD}Ready to start coding!${NC}"
    echo -e "${WHITE}  • Connect to remote: ${BOLD}git remote add origin <your-repo-url>${NC}"
    echo -e "${WHITE}  • Push to remote: ${BOLD}git push -u origin main${NC}"
    echo -e "${WHITE}  • Start developing your ${BOLD}$project_name${NC} project!\n"
    
    print_success "Happy coding! 🚀"
}

# Function to show help
show_help() {
    print_logo
    echo -e "${WHITE}${BOLD}Available commands:${NC}\n"
    echo -e "${CYAN}${BOLD}new${NC}     - Create a new Go project using Raidark as template"
    echo -e "${CYAN}${BOLD}help${NC}    - Show this help message"
    echo -e ""
    echo -e "${WHITE}${BOLD}Usage:${NC}"
    echo -e "${WHITE}  ./raidark.sh new${NC}     - Create new project from Raidark template"
    echo -e "${WHITE}  ./raidark.sh help${NC}    - Show help information"
    echo -e ""
}

# Main function
main() {
    case "${1:-help}" in
        "new")
            print_logo
            new_project
            ;;
        "help"|"--help"|"-h")
            show_help
            ;;
        *)
            print_logo
            print_error "Unknown command: $1"
            echo -e ""
            show_help
            exit 1
            ;;
    esac
}

# Check if script is being sourced or executed
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi 