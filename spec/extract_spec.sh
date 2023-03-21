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


It 'extract (with no args)'
  When call bin/csminv extract
  The status should equal 1
  # https://github.com/shellspec/shellspec/issues/295
  # The stderr should equal the contents of file "testdata/extract/extract_no_args.stderr"
  The lines of stderr should equal 9
  The line 1 of stderr should equal "Error: requires at least 1 arg(s), only received 0"
  The line 3 of stderr should equal "  csminv extract [...FILES] [flags]"
End

It 'extract --help'
  When call bin/csminv extract --help
  The status should equal 0
  The lines of stdout should equal 102
  The line 3 of stdout should equal "FIELD                                   FLAG                                     ENV                                     DEFAULT"
End

It 'extract testdata/fixtures/sls/sls_input_file_valid.json'
  When call bin/csminv extract testdata/fixtures/sls/sls_input_file_valid.json
  The status should equal 0
  The lines of stdout should equal 188
  The line 121 of stdout should equal '    "x3000": {'
  The line 161 of stdout should equal '        "FullName": "HMN Management Network Infrastructure",'
End

It 'extract testdata/fixtures/canu/paddle_valid.json'
  When call bin/csminv extract testdata/fixtures/canu/paddle_valid.json
  The status should equal 0
  The lines of stdout should equal 359
  The line 6 of stdout should equal '   "shcd_file": "Full_Architecture_Golden_Config_1.1.5.xlsx",'
  The line 13 of stdout should equal '     "model": "8325_JL627A",'
End

It 'extract testdata/fixtures/csi/system_config_valid.yaml'
  When call bin/csminv extract testdata/fixtures/csi/system_config_valid.yaml    
  The status should equal 0
  The lines of stdout should equal 129
  The line 49 of stdout should equal '   "hmn_dynamic_pool": "127.127.127.127/24",'
  The line 107 of stdout should equal '   "system_name": "seymour",'
End

It 'extract testdata/fixtures/csi/system_config_valid.yaml testdata/fixtures/canu/paddle_valid.json testdata/fixtures/sls/sls_input_file_valid.json'
  When call bin/csminv extract testdata/fixtures/csi/system_config_valid.yaml testdata/fixtures/canu/paddle_valid.json testdata/fixtures/sls/sls_input_file_valid.json 
  The status should equal 0
  The lines of stdout should equal 416
  The line 5 of stdout should equal '   "architecture": "network_v2",'
  The line 252 of stdout should equal '   "cabinets_yaml": "",'
  The line 393 of stdout should equal '          "Comment": "x3000c0h12s1",'
End

