import sbt._
import Keys._

name := "goldmane"
organization := "io.tigera"
version := "0.1"
scalaVersion := "3.3.5"
evictionErrorLevel := Level.Info
resolvers += Resolver.mavenLocal

libraryDependencies := Seq(
  "com.thesamet.scalapb" %% "scalapb-runtime-grpc" % "1.0.0-alpha.1",
  "com.thesamet.scalapb" %% "scalapb-runtime" % "1.0.0-alpha.1",
  "com.google.api-client" % "google-api-client-protobuf" % "2.7.2",
  "com.google.protobuf" % "protobuf-java" % "4.30.1",
  "io.grpc" % "grpc-protobuf" % "1.68.0",
  "io.grpc" % "grpc-stub" % "1.68.0",
  "com.thesamet.scalapb.common-protos" % "proto-google-common-protos-scalapb_0.11_3" % "2.9.6-0"
)