#!/bin/bash

# Chexum Project Cleanup Script
# Removes temporary build artifacts and frees up temporary storage space

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default options
VERBOSE=false
DRY_RUN=false
FORCE=false
THRESHOLD=80.0

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Function to show usage
show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Cleanup temporary files and build artifacts for the Chexum project.

OPTIONS:
    -v, --verbose       Enable verbose output
    -d, --dry-run       Show what would be cleaned without removing files
    -f, --force         Force cleanup even if storage usage is below threshold
    -t, --threshold N   Set storage usage threshold percentage (default: 80.0)
    -h, --help          Show this help message

EXAMPLES:
    $0                  # Clean if storage usage > 80%
    $0 -v -f            # Force cleanup with verbose output
    $0 -d               # Dry run to see what would be cleaned
    $0 -t 50            # Clean if storage usage > 50%

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -d|--dry-run)
            DRY_RUN=true
            shift
            ;;
        -f|--force)
            FORCE=true
            shift
            ;;
        -t|--threshold)
            THRESHOLD="$2"
            shift 2
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Check if Go cleanup command exists
CLEANUP_CMD="$PROJECT_ROOT/cmd/cleanup"
if [[ ! -f "$CLEANUP_CMD/main.go" ]]; then
    print_error "Cleanup command not found at $CLEANUP_CMD/main.go"
    exit 1
fi

# Build cleanup command arguments
CLEANUP_ARGS=""
if [[ "$VERBOSE" == "true" ]]; then
    CLEANUP_ARGS="$CLEANUP_ARGS -verbose"
fi
if [[ "$DRY_RUN" == "true" ]]; then
    CLEANUP_ARGS="$CLEANUP_ARGS -dry-run"
fi
if [[ "$FORCE" == "true" ]]; then
    CLEANUP_ARGS="$CLEANUP_ARGS -force"
fi
CLEANUP_ARGS="$CLEANUP_ARGS -threshold $THRESHOLD"

print_status "Starting cleanup with threshold: ${THRESHOLD}%"

# Change to project root
cd "$PROJECT_ROOT"

# Identify temp directory
TMP_DIR="${TMPDIR:-/tmp}"

# Check current usage
if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" ]]; then
    STORAGE_USAGE=$(df . | awk 'NR==2 {print $5}' | sed 's/%//')
else
    STORAGE_USAGE=$(df "${TMP_DIR}" | awk 'NR==2 {print $5}' | sed 's/%//')
fi
print_status "Current storage usage: ${STORAGE_USAGE}%"

# Run the Go cleanup command
if [[ "$DRY_RUN" == "true" ]]; then
    print_status "Running in dry-run mode..."
fi

print_status "Executing: go run $CLEANUP_CMD/main.go $CLEANUP_ARGS"

if go run "$CLEANUP_CMD/main.go" $CLEANUP_ARGS; then
    print_success "Cleanup completed successfully!"
else
    print_error "Cleanup failed!"
    exit 1
fi

# Show final usage
if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" ]]; then
    FINAL_USAGE=$(df . | awk 'NR==2 {print $5}' | sed 's/%//')
else
    FINAL_USAGE=$(df "${TMP_DIR}" | awk 'NR==2 {print $5}' | sed 's/%//')
fi
print_status "Final storage usage: ${FINAL_USAGE}%"

if [[ "$FINAL_USAGE" -lt "$STORAGE_USAGE" ]]; then
    SAVED=$((STORAGE_USAGE - FINAL_USAGE))
    print_success "Freed ${SAVED}% of storage space!"
fi
