#!/bin/bash

test_name=$(basename "$0" .sh)
test_out=out/tests/$test_name

mkdir -p "$test_out"

cat << EOF | $CC -o "$test_out"/a.o -c -xc -
#include <stdio.h>

int main(void) {
  printf("Hello, World\n");
  return 0;
}
EOF

$CC -B. -static "$test_out"/a.o -o "$test_out"/out
file "$test_out"/out | grep -q "ELF"