#!/bin/bash
# Generate Go code from ANTLR grammar files

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUTPUT_DIR="${SCRIPT_DIR}/../parser"

# Create output directory if not exists
mkdir -p "${OUTPUT_DIR}"

# Generate Go code
antlr4 -Dlanguage=Go -visitor -listener -package parser -o "${OUTPUT_DIR}" \
    "${SCRIPT_DIR}/SQLLexer.g4" \
    "${SCRIPT_DIR}/SQLParser.g4"

echo "Generated Go parser code in ${OUTPUT_DIR}"
