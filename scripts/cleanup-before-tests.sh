#!/bin/bash
# cleanup-before-tests.sh
# This script should be run BEFORE executing `go test` to ensure temporary storage has sufficient space
# It removes accumulated temporary files and Go build artifacts safely

set -e

# Identify the system temporary directory
TMP_DIR="${TMPDIR:-/tmp}"

# Default patterns to clean (relative to TMP_DIR)
PATTERNS=(
    "go-build*"
    "chexum-*"
    "checkpoint-*"
    "test-*"
)

echo "Cleaning up ${TMP_DIR} to prevent disk space issues during testing..."

# More aggressive cleanup using find with delete option
for pat in "${PATTERNS[@]}"; do
    # Use find with -delete for safer, more efficient removal
    find "${TMP_DIR}" -maxdepth 1 -name "$pat" -type d -exec rm -rf {} + 2>/dev/null || true
    find "${TMP_DIR}" -maxdepth 1 -name "$pat" -type f -delete 2>/dev/null || true
done

# Additional cleanup: remove any orphaned Go build directories
if [ -d "${TMP_DIR}" ]; then
    find "${TMP_DIR}" -maxdepth 1 -type d -name "go-*" -mmin +60 -exec rm -rf {} + 2>/dev/null || true
fi

echo "Cleanup complete. Current storage usage:"
if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" ]]; then
    df -h . | awk 'NR==2 {print $1 "\t" $5}'
else
    df -h "${TMP_DIR}" | awk 'NR==2 {print $1 "\t" $5}'
fi
