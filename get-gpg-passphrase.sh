#!/bin/bash
# This script determines where to read the GPG passphrase from.
# It will fail if no passphrase is available.

GPG_PASS=""

if command -v security >/dev/null 2>&1; then
  GPG_PASS=$(security find-generic-password -a "$USER" -s "gpg-passphrase" -w 2>/dev/null || echo "")
  [ -n "$GPG_PASS" ] && echo "  • ✅  Using GPG passphrase from MacOS Keychain" >&2
fi

if [ -z "$GPG_PASS" ] && [ -n "$GPG_PASSPHRASE" ]; then
  GPG_PASS="$GPG_PASSPHRASE"
  echo "  • ✅  Using GPG passphrase from environment variable" >&2
fi

if [ -n "$GPG_PASS" ]; then
  GPG_PASSPHRASE="$GPG_PASS"
  export GPG_PASSPHRASE
else
  echo "  • ❌  No GPG passphrase found!" >&2
  echo "" >&2
  echo "Please set the passphrase using one of these methods:" >&2
  if command -v security >/dev/null 2>&1; then
    echo "   * Store in MacOS Keychain:" >&2
    echo "     security add-generic-password -a \"\$USER\" -s \"gpg-passphrase\" -w \"your_passphrase\"" >&2
  fi
  echo "   * Set environment variable:" >&2
  echo "     export GPG_PASSPHRASE=\"your_passphrase\"" >&2
  echo "" >&2
  exit 1
fi