#!/usr/bin/env bash
# Copyright (c) Adfinis
# SPDX-License-Identifier: GPL-3.0-or-later


docker compose exec -T bastion \
    /opt/bastion/bin/admin/setup-first-admin-account.sh \
    bastionadmin \
    auto < ssh-keys/id_ed25519.pub

# enable ipv6 support
docker compose exec -T bastion \
    /usr/bin/sed -i 's/"IPv6Allowed": false,/"IPv6Allowed": true,/' \
    /etc/bastion/bastion.conf
