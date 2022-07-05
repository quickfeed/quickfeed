#image/gradle:5.6-jdk12

echo "\n=== Preparing for Test Execution ===\n"

ping -c 4 google.com 2>&1
ls

ASSIGNDIR=/quickfeed/assignments/{{ .AssignmentName }}/
TESTDIR=/quickfeed/tests/{{ .AssignmentName }}/

cat <<EOF> /home/gradle/.gradle/gradle.properties
org.gradle.parallel=true
org.gradle.daemon=true
org.gradle.jvmargs=-Xms256m -Xmx1024m
EOF

# Make sure there are not tests in the student repo
rm -rf $ASSIGNDIR/src/test/*

echo "Removed tests folder on user file\n"

# Generate new Secret.java with new secret value for each run
cd test
cat <<EOF > $TESTDIR/src/test/java/common/SecretClass.java
package common;

public class SecretClass {
public static String getSecret() {
  return "{{ .RandomSecret }}";
}
}
EOF



# Fail student code that attempts to access secret
#cd $ASSIGNDIR/
#if grep --quiet -r -e common.Secret -e GlobalSecret * ; then
#  echo "\n=== Misbehavior Detected: Failed ===\n"
#  exit
#fi

# Copy tests into student assignments folder for running tests
cp -r $TESTDIR/src/test/* $ASSIGNDIR/src/test/
echo "copied test files to user folder \n"
echo "$ASSIGNDIR/src/test/"
echo `ls $ASSIGNDIR/src/test/`

cp $TESTDIR/build.gradle $ASSIGNDIR/build.gradle
cp $TESTDIR/gradlew $ASSIGNDIR/gradlew
cd $ASSIGNDIR/

# Perform lab specific setup
if [ -f "setup.sh" ]; then
    bash setup.sh
fi

echo "\n=== Running Tests ===\n"
gradle clean test 2>&1
echo "\n=== Finished Running Tests ===\n"
