# List of TODOs and open questions

This list is outdated and will be removed, when any remaining items have been added to Trello or the Issue tracker.

## Concerns

* Rethink how assignments work and how they are fetched.
* How to choose what to test, could a student write such a file with such a test?
* How much effort should be put into the course pipeline for making cheating difficult?
* Where to store information about approved submission, in a new table, or on the latest submission?

## Other things to think about

* How the CI system should be interfaced.
* Should the client check with the server for the current CI status at a regular interval?
  * Some running tasks system?
  * What happens if more then one CI runs at a time?
  * Where should the score be calculated (Hein: we calculate in backend.)

## Remaining unit MVP

* Refactor some code inside hooks, a little redundant code.
* Secret handling, or some better way of output sorting and stuff.
* Make test cases with integration testing.
* Update readme with information about webhooks on providers, what providers are supported.
  * need ssl sertificat fix
  * How to set up providers
