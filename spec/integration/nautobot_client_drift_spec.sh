#!/usr/bin/env sh
#
# MIT License
#
# (C) Copyright 2026 Hewlett Packard Enterprise Development LP
#
# Permission is hereby granted, free of charge, to any person obtaining a
# copy of this software and associated documentation files (the "Software"),
# to deal in the Software without restriction, including without limitation
# the rights to use, copy, modify, merge, publish, distribute, sublicense,
# and/or sell copies of the Software, and to permit persons to whom the
# Software is furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included
# in all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
# THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
# OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
# ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
# OTHER DEALINGS IN THE SOFTWARE.
#

# Codegen check: the committed Nautobot client matches the pinned simulator.
#
# This is a generated-artifact freshness gate, not a product-integration test:
# it never exercises the cani binary, it only regenerates the client and diffs
# it against the committed source. It lives under spec/integration/ solely
# because that is the only tier with a live Nautobot to source the OpenAPI spec.
#
# Prerequisites:
#   - Nautobot running locally (make nautobot-up)
#
# Unlike the other external tests it ignores SKIP_EXTERNAL_TESTS (it needs no
# seed data) and skips only when Nautobot is unreachable.

Describe 'CODEGEN: Nautobot generated client drift'

  # This check is intentionally NOT gated on SKIP_EXTERNAL_TESTS: it needs no
  # seed data and only compares the committed client to the live simulator, so
  # it should run whenever Nautobot is reachable. It skips gracefully otherwise.
  # The reachability probe must be a function: `Skip if 'x' ! cmd` is silently
  # ignored by shellspec because `!` is a shell keyword, not a command.
  nautobot_unreachable() {
    ! curl -sf -H "Authorization: Token ${NAUTOBOT_TOKEN}" \
      "${NAUTOBOT_URL}/status/" >/dev/null 2>&1
  }

  Skip if 'Nautobot is not reachable' nautobot_unreachable

  # Regenerate the client from the live simulator and compare it against the
  # committed source. Runs from the repository root so make can find its targets.
  # NAUTOBOT_URL is unset here because the suite exports it with an /api suffix,
  # while the Makefile expects the base URL and appends /api/... itself.
  check_client_drift() {
    ( cd "$SHELLSPEC_HELPERDIR/.." && env -u NAUTOBOT_URL make nautobot_client_check )
  }

  It 'has no drift against the committed client'
    When call check_client_drift
    The status should equal 0
    The output should include 'Nautobot client check passed'
  End

End
