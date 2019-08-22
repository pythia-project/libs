#!/bin/sh
javac -classpath json-20180813.jar:commons-csv-1.7.jar:. -Xlint org/pythia/Execute.java
jar cmf MANIFEST.mf pythia-1.0.jar org/pythia/*.class
