# Java build

The java build uses the gradle build system and contains a build.gradle with the default build information that is needed build a project.

It is also modified to use the following folder structure when it builds the project:

## Folder structure for one assignment
```
rootfolder (project folder)
    src (src root, needs to be named src) 
        package1 (java package)
            javafile.java
            javafile2.java
        package2 (java package)
            javafile3.java
            javafile4.java
    test (test root, needs to be named test)
        testpackage1 (java package)
            testjavafile.java
            testjavafile2.java
        testpackage2 (java package)
            testjavafile3.java
            testjavafile4.java
    build.gradle
```

## Folder structure for assignments, tests, and solutions
The assignment repo should look like this
```
assignments (repo)
    oving1 (project/assignment)
        src (source)
            package1 (java package)
                file1.java (java file)
                file2.java
            package2
                file3.java
                file4.java
    oving2 (project/assignment)
        src (source)
            package1 (java package)
                file1.java
                file2.java
            package2
                file3.java
                file4.java
```
Tests repository should look like this
```
tests (repo)
    oving1 (project/assignment)
        test (tests)
            package1 (java package)
                file1.java (java file)
                file2.java
            package2
                file3.java
                file4.java
    oving2 (project/assignment)
        test (tests)
            package1 (java package)
                file1.java
                file2.java
            package2
                file3.java
                file4.java
```

The solutions repo should have the same folderstructure as assignments

## Shell files
### runcontainer.sh: `usage: runcontainer.sh username baseurl`

This is the mainfile to run the docker container. username is the name of the user to clone the repository for, and the baseurl is the url to the organisation. 

### testjava.sh: `usage: testjava.sh username baseurl`

Is the bootstraping file sent to the container, to clone, create build.gradle files, and invoke junit inside the running container

### build.gradle

Is the default build file, and it is also embedded inside the testjava.sh file, to be created if it does not exist already inside the java project folder.