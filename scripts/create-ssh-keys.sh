#!/usr/bin/env bash
# Copyright (c) Adfinis
# SPDX-License-Identifier: GPL-3.0-or-later


mkdir -p ssh-keys
ssh-keygen -t ed25519 -f ssh-keys/id_ed25519 -N "" -C "bastion-test-key"

cp ssh-keys/* internal/provider/testutils/testdata/ssh-keys/
