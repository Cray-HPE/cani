#!/usr/bin/env sh
#
# MIT License
#
# (C) Copyright 2023-2026 Hewlett Packard Enterprise Development LP
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
Describe 'INTEGRATION:'

It 'migrates a v1alpha1 datastore to v1alpha3'
	BeforeCall 'setup_datastore_migration_env canitestdb_v1alpha1.json'
	When call bin/cani alpha show --config "$CANI_CONF"
	The status should equal 0
	The stdout should be present
	The stderr should include 'creating default config'
	The stderr should include 'Migrated datastore from v1alpha1 to v1alpha3'
	The file "$CANI_CONF" should be exist
	The file "$CANI_DS" should be exist
	The file "$CANI_DS.canisave" should be exist
	The contents of file "$CANI_DS" should include '"schemaVersion"'
	The contents of file "$CANI_DS" should include '"v1alpha3"'
	The contents of file "$CANI_DS" should include '"devices"'
	The contents of file "$CANI_DS" should include '"locations"'
	The contents of file "$CANI_DS" should include '"racks"'
	The contents of file "$CANI_DS" should not include '"Hardware"'
	The contents of file "$CANI_DS" should not include '"v1alpha1"'
End

End
