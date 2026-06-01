#!/bin/bash

# ARE Installation Script
# Installs ARE binary and downloads stdlib from GitHub
# Auto-detects Linux or macOS

set -e

VERSION=${1:-"0.2.12"}
BINARY_PATH=${2:-""}
GITHUB_REPO="ECMA-King87/ArachnoScript-Runtime-Engine-ARE-"
GITHUB_BASE_URL="https://raw.githubusercontent.com/$GITHUB_REPO/main/versions"
GITHUB_RELEASES_URL="https://github.com/$GITHUB_REPO/releases/download"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Detect OS and architecture
detect_platform() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        PLATFORM="linux"
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        PLATFORM="macos"
    else
        echo -e "${RED}Error: Unsupported OS: $OSTYPE${NC}"
        exit 1
    fi
    
    ARCH=$(uname -m)
    if [ "$ARCH" = "x86_64" ]; then
        ARCH="amd64"
    elif [ "$ARCH" = "aarch64" ]; then
        ARCH="arm64"
    fi
}

# Set installation paths
set_install_paths() {
    local BASE_DIR
    if [ "$EUID" -eq 0 ]; then
        BASE_DIR="/usr/local"
    else
        BASE_DIR="$HOME/.local"
    fi
    
    INSTALL_DIR="$BASE_DIR/bin"
    STDLIB_DIR="$BASE_DIR/share/are/stdlib"
    EXAMPLES_DIR="$BASE_DIR/share/are/examples"
    SRC_DIR="$BASE_DIR/share/are/src"
    
    if [ "$EUID" -eq 0 ]; then
        CONFIG_DIR="/etc/are"
    else
        CONFIG_DIR="$HOME/.config/are"
    fi
}

# Download binary from GitHub if needed
setup_binary() {
    # Use provided binary path if available
    if [ -n "$BINARY_PATH" ] && [ -f "$BINARY_PATH" ]; then
        return 0
    fi
    
    # Try to find locally built binary
    local local_binary="./build/$PLATFORM/are-$PLATFORM-$ARCH"
    if [ -f "$local_binary" ]; then
        BINARY_PATH="$local_binary"
        return 0
    fi
    
    # Download from GitHub releases
    echo -e "${YELLOW}Local binary not found, attempting to download from GitHub...${NC}"
    
    local download_dir=$(mktemp -d)
    local release_url="$GITHUB_RELEASES_URL/v$VERSION/are-$PLATFORM-$ARCH"
    
    echo -n "  Downloading are-$PLATFORM-$ARCH... "
    if curl -sfL -o "$download_dir/are-binary" "$release_url" 2>/dev/null; then
        chmod +x "$download_dir/are-binary"
        BINARY_PATH="$download_dir/are-binary"
        echo -e "${GREEN}✓${NC}"
    else
        echo -e "${RED}✗${NC}"
        echo -e "${RED}Error: Could not find or download binary${NC}"
        echo -e "${YELLOW}Options:${NC}"
        echo "  1. Build locally: bash scripts/build.sh"
        echo "  2. Check internet connection"
        echo "  3. Download manually from: $GITHUB_RELEASES_URL/v$VERSION/"
        rm -rf "$download_dir"
        exit 1
    fi
}

# Install binary
install_binary() {
    echo -e "${YELLOW}Installing binary...${NC}"
    mkdir -p "$INSTALL_DIR"
    cp "$BINARY_PATH" "$INSTALL_DIR/are"
    chmod +x "$INSTALL_DIR/are"
    
    # Handle macOS code signing
    if [ "$PLATFORM" = "macos" ] && [ "$EUID" -eq 0 ]; then
        xattr -d com.apple.quarantine "$INSTALL_DIR/are" 2>/dev/null || true
    fi
    
    echo -e "${GREEN}✓ Binary installed to $INSTALL_DIR/are${NC}"
}

# Setup PATH
setup_path() {
    echo -e "${YELLOW}Configuring PATH...${NC}"
    
    if [[ "$PATH" == *"$INSTALL_DIR"* ]]; then
        echo -e "${GREEN}✓ $INSTALL_DIR already in PATH${NC}"
        return
    fi
    
    if [ "$INSTALL_DIR" = "$HOME/.local/bin" ]; then
        # For user-local installation
        local shell_rc=""
        
        if [ -n "$ZSH_VERSION" ] || [ -f "$HOME/.zshrc" ]; then
            shell_rc="$HOME/.zshrc"
        elif [ -n "$BASH_VERSION" ] || [ -f "$HOME/.bashrc" ]; then
            shell_rc="$HOME/.bashrc"
        fi
        
        if [ -n "$shell_rc" ] && [ -f "$shell_rc" ]; then
            if ! grep -q "\.local/bin" "$shell_rc"; then
                echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$shell_rc"
                echo -e "${GREEN}✓ Added to PATH in $shell_rc${NC}"
            fi
        fi
    else
        # For system-wide installation
        if [ "$PLATFORM" = "linux" ] && [ -d "/etc/profile.d" ]; then
            cat > /etc/profile.d/are-path.sh << 'EOF'
if [ -d "/usr/local/bin" ]; then
    export PATH="/usr/local/bin:$PATH"
fi
EOF
            chmod 644 /etc/profile.d/are-path.sh
            echo -e "${GREEN}✓ Added to PATH via /etc/profile.d/are-path.sh${NC}"
        elif [ "$PLATFORM" = "macos" ] && [ -d "/etc/paths.d" ]; then
            echo "$INSTALL_DIR" | tee /etc/paths.d/are > /dev/null
            echo -e "${GREEN}✓ Added to PATH via /etc/paths.d/are${NC}"
        fi
    fi
}

# Download resources from GitHub
download_resources() {
    echo -e "\n${YELLOW}Downloading resources from GitHub v$VERSION...${NC}"
    
    local temp_dir=$(mktemp -d)
    trap "rm -rf $temp_dir" EXIT
    
    download_file() {
        local filename=$1
        local source_path=$2
        local dest_dir=$3
        
        local url="$GITHUB_BASE_URL/v$VERSION/$source_path/$filename"
        echo -n "  Downloading $filename... "
        
        if curl -sf -o "$temp_dir/$filename" "$url" 2>/dev/null; then
            mkdir -p "$dest_dir"
            mv "$temp_dir/$filename" "$dest_dir/$filename"
            chmod 644 "$dest_dir/$filename"
            echo -e "${GREEN}✓${NC}"
            return 0
        else
            echo -e "${YELLOW}⚠${NC}"
            return 1
        fi
    }
    
    # Download stdlib
    echo -e "\n${YELLOW}Stdlib modules:${NC}"
    local stdlib_modules=(
        "arrays" "date" "encoding" "fs" "http" "io"
        "math" "numbers" "objects" "promise" "regexp" "runtime"
        "strings" "symbols"
    )
    
    local stdlib_count=0
    for module in "${stdlib_modules[@]}"; do
        if download_file "$module.as" "stdlib" "$STDLIB_DIR"; then
            ((stdlib_count++))
        fi
    done
    
    # Download examples
    echo -e "\n${YELLOW}Examples:${NC}"
    local examples=("http.as" "num-sorter.as" "ffi.as")
    local examples_count=0
    for example in "${examples[@]}"; do
        if download_file "$example" "examples" "$EXAMPLES_DIR"; then
            ((examples_count++))
        fi
    done
    
    # Download source files
    echo -e "\n${YELLOW}Source files:${NC}"
    local source_files=("main.as" "smoke_c.as")
    local src_count=0
    for srcfile in "${source_files[@]}"; do
        if download_file "$srcfile" "src/as" "$SRC_DIR"; then
            ((src_count++))
        fi
    done
    
    echo ""
    echo -e "${GREEN}✓ Downloaded:${NC}"
    echo "  - $stdlib_count stdlib modules"
    [ $examples_count -gt 0 ] && echo "  - $examples_count example files"
    [ $src_count -gt 0 ] && echo "  - $src_count source files"
}

# Main installation
main() {
    echo -e "${BLUE}=== ARE Installation ===${NC}"
    
    detect_platform
    echo "Platform: $PLATFORM ($ARCH)"
    echo "Version: $VERSION"
    echo ""
    
    set_install_paths
    
    echo -e "${YELLOW}Installation directories:${NC}"
    echo "  Binary:   $INSTALL_DIR/are"
    echo "  Stdlib:   $STDLIB_DIR"
    echo "  Examples: $EXAMPLES_DIR"
    echo "  Config:   $CONFIG_DIR"
    echo ""
    
    # Create directories
    mkdir -p "$INSTALL_DIR" "$STDLIB_DIR" "$EXAMPLES_DIR" "$SRC_DIR" "$CONFIG_DIR"
    
    # Setup and install binary
    setup_binary
    install_binary
    
    # Verify binary works
    echo -e "${YELLOW}Verifying installation...${NC}"
    if "$INSTALL_DIR/are" --version > /dev/null 2>&1 || true; then
        echo -e "${GREEN}✓ Binary verified${NC}"
    else
        echo -e "${YELLOW}⚠ Could not verify binary (this may be normal)${NC}"
    fi
    
    # Setup PATH
    setup_path
    
    # Download resources
    download_resources
    
    # Create version file
    echo "$VERSION" > "$CONFIG_DIR/version"
    echo -e "${GREEN}✓ Version file created${NC}"
    
    # Summary
    echo -e "\n${GREEN}=== Installation Complete ===${NC}"
    echo ""
    echo -e "ARE v$VERSION has been installed!"
    echo ""
    echo -e "${YELLOW}Next steps:${NC}"
    if [ "$INSTALL_DIR" = "$HOME/.local/bin" ]; then
        echo "  1. Reload your shell: source ~/.bashrc (or ~/.zshrc)"
    else
        echo "  1. Open a new terminal or run: exec \$SHELL"
    fi
    echo "  2. Test: are --version"
    echo "  3. Run: are script.as"
    echo ""
    echo -e "${YELLOW}Locations:${NC}"
    echo "  Binary:   $INSTALL_DIR/are"
    echo "  Stdlib:   $STDLIB_DIR"
    echo "  Examples: $EXAMPLES_DIR"
    echo "  Source:   $SRC_DIR"
    echo ""
    echo -e "${YELLOW}Help:${NC}"
    echo "  are --help"
    echo "  https://github.com/$GITHUB_REPO"
}

main "$@"
