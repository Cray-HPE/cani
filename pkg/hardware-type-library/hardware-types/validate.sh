#! /usr/bin/env bash

failed=0

for file in *.yaml; do 
    echo
    echo "========================================"
    echo "Validating ${file}"
    echo "========================================"
    if ! jv -output detailed schema/devicetype.json "$file"; then
        echo "Not Valid"
        failed=1
    fi   
done

echo "========================================"
echo "Validation Results"
echo "========================================"

if [[ $failed -eq 1 ]]; then
    echo "One or more hardware types have validation errors"
    exit 1
else
    echo "Pass"
fi