#!/usr/bin/env sh
# MIT License
#
# (C) Copyright 2023 Hewlett Packard Enterprise Development LP
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

It 'add cabinet (with no args)'
  When call bin/cani add cabinet
  The status should equal 0
  The lines of stdout should equal 7
  The line 1 of stdout should equal "Unpacking add_liquid_cooled_cabinet.py..."
  The line 2 of stdout should equal "Unpacking backup_sls_postgres.sh..."
  The line 3 of stdout should equal "Unpacking inspect_sls_cabinets.py..."
  The line 4 of stdout should equal "Unpacking update-ncn-cabinet-routes.sh..."
  The line 5 of stdout should equal "Unpacking update_ncn_etc_hosts.py..."
  The line 6 of stdout should equal "Unpacking verify_bmc_credentials.sh..."
End

It '--debug add cabinet'
  When call bin/cani --debug add cabinet
  The status should equal 0
  The lines of stdout should equal 7
  The line 1 of stdout should equal "Unpacking add_liquid_cooled_cabinet.py..."
  The line 2 of stdout should equal "Unpacking backup_sls_postgres.sh..."
  The line 3 of stdout should equal "Unpacking inspect_sls_cabinets.py..."
  The line 4 of stdout should equal "Unpacking update-ncn-cabinet-routes.sh..."
  The line 5 of stdout should equal "Unpacking update_ncn_etc_hosts.py..."
  The line 6 of stdout should equal "Unpacking verify_bmc_credentials.sh..."
End

It '--debug add cabinet cabinet1'
  When call bin/cani --debug add cabinet cabinet1
  The status should equal 0
  The lines of stdout should equal 7
  The line 1 of stdout should equal "Unpacking add_liquid_cooled_cabinet.py..."
  The line 2 of stdout should equal "Unpacking backup_sls_postgres.sh..."
  The line 3 of stdout should equal "Unpacking inspect_sls_cabinets.py..."
  The line 4 of stdout should equal "Unpacking update-ncn-cabinet-routes.sh..."
  The line 5 of stdout should equal "Unpacking update_ncn_etc_hosts.py..."
  The line 6 of stdout should equal "Unpacking verify_bmc_credentials.sh..."
  The line 7 of stdout should equal 'add cabinet called'
  The lines of stderr should equal 1
  The stderr should include '{"level":"debug","time":'
  The stderr should include '"message":"Added cabinet cabinet1"}'
End
