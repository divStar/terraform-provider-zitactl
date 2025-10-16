#!/bin/bash
# Copyright (c) Igor Voronin
# SPDX-License-Identifier: MIT

# This script unlocks the GPG key in gpg-agent using a passphrase from
# environment variable or MacOS Keychain.

GPG_PASS=""

if command -v security >/dev/null 2>&1; then
  GPG_PASS=$(security find-generic-password -a "$USER" -s "gpg-passphrase" -w 2>/dev/null || echo "")
  [ -n "$GPG_PASS" ] && echo "  • ✅  Using GPG passphrase from MacOS Keychain"
fi

if [ -z "$GPG_PASS" ] && [ -n "$GPG_PASSPHRASE" ]; then
  GPG_PASS="$GPG_PASSPHRASE"
  echo "  • ✅  Using GPG passphrase from environment variable"
fi

if [ -z "$GPG_PASS" ]; then
  echo "  • ❌  No GPG passphrase found!"
  echo ""
  echo "Please set the passphrase using one of these methods:"
  if command -v security >/dev/null 2>&1; then
    echo "   * Store in MacOS Keychain:"
    echo "     security add-generic-password -a \"\$USER\" -s \"gpg-passphrase\" -w \"your_passphrase\""
  fi
  echo "   * Set environment variable:"
  echo "     export GPG_PASSPHRASE=\"your_passphrase\""
  echo ""
  exit 1
fi

# Unlock the key in gpg-agent by performing a test sign operation
echo "  • 🔐  Unlocking GPG key in gpg-agent..."
if gpg --batch --yes --pinentry-mode loopback --passphrase "$GPG_PASS" --sign --local-user "$GPG_FINGERPRINT" --armor </dev/null >/dev/null 2>&1; then
  echo "  • ✅  GPG key unlocked and cached in gpg-agent"
else
  echo "  • ❌  Failed to unlock GPG key"
  exit 1
fi