## Conserns

* Rethink how assignments work and how they are fetched, What to choose what to test, could a student write such a file with such a test?
* How much effort should be put into the course pipeline for making cheating difficult
* Where to store information about approved submission, in a new table, or on the latest submission?
* What to run CI on webhook? Run on the entier soulution, or just on the latest active 
* Store information about git repo location, so there is no need to ask github/gitlab each time
* 


## Other things to think about
* Extract all of information to run the ci to its own struct.
* Move script to external script which are read on build, which makes them easier to edit and customize. 
* Maybe use json/yaml to configure the different languages.
* How the CI system should be interfaced, Should the client check with the server for the current status, at an interval?
    * Some running tasks system?
    * What happens if more then on CI runs at a time?
    * Where should the score be calculated
    * 

## Reminding unit MVP
* Creating repositories with priviligies,
* Checking to see if a organisation/directory allows private repos, or not
* Be able to create private repositories. 
* Refactor some code inside hooks, a little redundant code
* Figure out way on how to approve an assingment
* Running CI only on the correct assignment
* Secret handling, or some better way of output sorting and stuff.
* Need organisation with private repos, send message to Hein.
* Make test cases with integration testing.
* Update readme with information about webhooks on providers, what providers are supported, 
    * how to setup docker, 
    * need ssl sertificat fix
    * How to set up providers
    * 