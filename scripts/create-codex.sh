#!/bin/bash

# Raidark Codex File Generator
# This script creates AGENTS.md file for Codex tool by concatenating .cursor/rules/*.mdc files

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
RULES_DIR=".cursor/rules"
OUTPUT_FILE="AGENTS.md"
GITIGNORE_FILE=".gitignore"

# Function to display usage
show_usage() {
    echo -e "${BLUE}Raidark Codex File Generator${NC}"
    echo ""
    echo "Usage: $0 codex [options]"
    echo ""
    echo "Options:"
    echo "  --output FILE    Specify output file (default: AGENTS.md)"
    echo "  --force          Overwrite existing file without confirmation"
    echo "  --preview        Show what would be generated without creating file"
    echo ""
    echo "This command will:"
    echo "  - Read all .mdc files from .cursor/rules/"
    echo "  - Concatenate them into a single AGENTS.md file"
    echo "  - Add AGENTS.md to .gitignore if not already present"
    echo "  - Create context file for Codex tool"
    echo ""
    echo "Example:"
    echo "  $0 codex"
    echo "  $0 codex --preview"
    echo "  $0 codex --output MY_AGENTS.md"
}

# Function to check if rules directory exists
check_rules_directory() {
    if [ ! -d "$RULES_DIR" ]; then
        echo -e "${RED}Error: Rules directory not found at $RULES_DIR${NC}"
        echo -e "${YELLOW}Expected directory structure: .cursor/rules/${NC}"
        exit 1
    fi
}

# Function to find .mdc files
find_mdc_files() {
    find "$RULES_DIR" -name "*.mdc" -type f | sort
}

# Function to check if there are .mdc files
check_mdc_files() {
    local mdc_files=$(find_mdc_files)
    
    if [ -z "$mdc_files" ]; then
        echo -e "${RED}Error: No .mdc files found in $RULES_DIR${NC}"
        echo -e "${YELLOW}Expected files like: commit.mdc, devmode.mdc, etc.${NC}"
        exit 1
    fi
    
    echo -e "${BLUE}Found .mdc files:${NC}"
    echo "$mdc_files" | while read -r file; do
        echo -e "${YELLOW}  - $file${NC}"
    done
    echo ""
}

# Function to generate header for AGENTS.md
generate_header() {
    cat << EOF
# AGENTS.md - Codex Context File

This file contains the concatenated content of all .mdc files from .cursor/rules/
Generated on: $(date)

> **Note**: This file is auto-generated. Do not edit manually.
> To update, run: ./raidark.sh codex

---

EOF
}

# Function to process single .mdc file
process_mdc_file() {
    local file="$1"
    local filename=$(basename "$file")
    
    echo "## $filename"
    echo ""
    
    # Add file content
    cat "$file"
    echo ""
    echo "---"
    echo ""
}

# Function to generate AGENTS.md content
generate_agents_content() {
    local mdc_files=$(find_mdc_files)
    
    # Generate header
    generate_header
    
    # Process each .mdc file
    echo "$mdc_files" | while read -r file; do
        process_mdc_file "$file"
    done
}

# Function to preview content
preview_content() {
    echo -e "${BLUE}Preview of AGENTS.md content:${NC}"
    echo ""
    echo "=========================================="
    generate_agents_content
    echo "=========================================="
    echo ""
    echo -e "${YELLOW}Preview completed. Use 'codex' without --preview to create the file.${NC}"
}

# Function to update .gitignore
update_gitignore() {
    local output_file="$1"
    
    if [ -f "$GITIGNORE_FILE" ]; then
        # Check if the file is already in .gitignore
        if grep -q "^$output_file$" "$GITIGNORE_FILE"; then
            echo -e "${YELLOW}$output_file already in .gitignore${NC}"
        else
            echo -e "${BLUE}Adding $output_file to .gitignore${NC}"
            echo "$output_file" >> "$GITIGNORE_FILE"
            echo -e "${GREEN}Added $output_file to .gitignore${NC}"
        fi
    else
        echo -e "${YELLOW}Warning: .gitignore not found${NC}"
        echo -e "${BLUE}Creating .gitignore with $output_file${NC}"
        echo "$output_file" > "$GITIGNORE_FILE"
        echo -e "${GREEN}Created .gitignore with $output_file${NC}"
    fi
}

# Function to create AGENTS.md file
create_agents_file() {
    local output_file="$1"
    local force="$2"
    
    # Check if file exists and handle accordingly
    if [ -f "$output_file" ] && [ "$force" = "false" ]; then
        echo -e "${YELLOW}Warning: $output_file already exists${NC}"
        read -p "Do you want to overwrite it? (y/N): " -n 1 -r
        echo
        
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo -e "${RED}Operation cancelled${NC}"
            exit 1
        fi
    fi
    
    echo -e "${BLUE}Generating $output_file...${NC}"
    
    # Generate content and write to file
    generate_agents_content > "$output_file"
    
    echo -e "${GREEN}✓ $output_file created successfully${NC}"
    
    # Update .gitignore
    update_gitignore "$output_file"
    
    # Show file statistics
    local line_count=$(wc -l < "$output_file")
    local word_count=$(wc -w < "$output_file")
    local size=$(du -h "$output_file" | cut -f1)
    
    echo ""
    echo -e "${BLUE}File Statistics:${NC}"
    echo -e "${YELLOW}  Lines: $line_count${NC}"
    echo -e "${YELLOW}  Words: $word_count${NC}"
    echo -e "${YELLOW}  Size: $size${NC}"
    
    echo ""
    echo -e "${GREEN}AGENTS.md file ready for Codex tool!${NC}"
    echo -e "${BLUE}You can now use this file as context for your Codex operations.${NC}"
}

# Main execution
main() {
    local output_file="$OUTPUT_FILE"
    local force=false
    local preview=false
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --output)
                output_file="$2"
                shift 2
                ;;
            --force)
                force=true
                shift
                ;;
            --preview)
                preview=true
                shift
                ;;
            codex)
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
    
    check_rules_directory
    check_mdc_files
    
    if [ "$preview" = "true" ]; then
        preview_content
    else
        create_agents_file "$output_file" "$force"
    fi
}

# Execute main function with all arguments
main "$@" 