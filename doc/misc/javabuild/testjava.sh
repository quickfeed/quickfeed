#!/bin/sh
# This is a bootstrap file that will clone and
# run all tests found the the merged folder.
# It will also create a default build.gradle file
# if it does not exist, but use the existing one,
# if it does.

# usage: testjava [username] [baserepo]

if [ -z "$1" ]; then
    NAME=""
else
    NAME=$1-
fi
if [ -z "$2" ]; then
    LINK="https://github.com/AutograderTestOrg1"
else
    LINK=$2
fi

MERGE="merged"
# Buildfile
BF="build.gradle"
AD="$NAME"assignments
echo $PWD
mkdir -p pull; cd pull
mkdir -p $MERGE

git clone $LINK/$AD
git clone $LINK/tests

cp -af $AD/. $MERGE/
cp -af tests/. $MERGE/
basePath=$PWD

# Check if all folders have the $BF file
for folder in $(ls $MERGE)
do
    curPath=$PWD/$MERGE/$folder
    if [ -d $curPath ]; then
        if [ -f $curPath/$BF ]; then
            echo "Found $BF in $curPath/$BF"
        else
            echo "Missing $BF in $curPath creating file"
            echo "apply plugin: 'java'

repositories {
    mavenCentral()
}

test { 
    testLogging.showStandardStreams = true
}

sourceCompatibility = 1.8
targetCompatibility = 1.8

sourceSets{
    main{ 
        java {
            srcDir 'src'
        }
    }
    test {
        java {
            srcDir 'test'
        }
    }
}

dependencies {
    testCompile 'junit:junit:4.12'
}
 
jar {
    baseName = 'ovigner' 
    version =  '0.1.0'
}" > $curPath/$BF
        fi
    fi
done

# Run the $BF file in all the folders
for folder in $(ls $MERGE)
do
    if [ -d $basePath/$MERGE/$folder ]; then
        curPath=$basePath/$MERGE/$folder
        cd $curPath
        echo "====Running tests for $folder===="
        gradle clean test
    fi
done