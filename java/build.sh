#!/bin/sh
javac -classpath json-20180813.jar:commons-csv-1.7.jar:. -Xlint org/pythia/Execute.java
if [ $? -eq 0 ]
then
	jar cmf MANIFEST.mf pythia-1.0.jar org/pythia/*.class
	if [ $? -eq 0 ]
	then
		echo "pythia-1.0.jar created.\nBuild succeeded!"
		exit 0
	fi
fi
echo "Build failed!"
