#!/bin/bash

test_name=$(basename "$0" .sh)
test_out=out/tests/$test_name

mkdir -p "$test_out"

cat << EOF | riscv64-linux-gnu-gcc -o "$test_out"/a.o -c -xc -
#include <stdio.h>

int main(void) {
  printf("Hello, World\n");
  return 0;
}
EOF

./rvld "$test_out"/a.o