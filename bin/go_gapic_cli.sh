#!/bin/bash
mkdir -p src/go/gapic
bazel build //tigera/goldmane/v1:goldmane_go_gapic_srcjar
unzip bazel-bin/tigera/goldmane/v1/goldmane_go_gapic_srcjar.srcjar \
         -d src/go/gapic
mv src/go/gapic/google.golang.org src/go/gapic/cloud.tigera.io
cp goldmane_v1/go/cli/* src/go