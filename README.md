requires: bazel 6.5, buildozer 6.4, jdk(openjdk-default is fine) and JAVA_HOME env var, python, g++, make, golang and build-essentials
1: bazel run //:build_gen -- --src=tigera
2: sed -i "s/\/\/google/@com_google_googleapis\/\/google/g" tigera/goldmane/v1/BUILD.bazel
3: bazel query ...:*
4: bazel build //tigera/goldmane/v1:goldmane_py_gapic
