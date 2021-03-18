
### `npm install`

### `webpack`

### Things related to functionality yet to be solved:
#### State
Doesn't wait for state or it doesn't update on state change. Declaring local states might solve this issue, mby.
Otherwise maybe a smart choice for dependency array. But that suually just results in the useEffect() method being ran twice and components.
```
    useEffect(() => {
        actions.getEnrollmentsByUser()
        .then(success => {
            if (success) {
                state.enrollments.map(enrol =>{
                    actions.getSubmissions(enrol.getCourseid())
                    actions.getAssignmentsByCourse(enrol.getCourseid()).then(() =>{
                        console.log(state.assignments)
                    })
                    
                })
            }
        });
    }, []) <----
  
```
Here at the last line the empty array indicates that this will only run once, when the component mounts. So state changes doesn't come through, unless we
We choose a `[dependency array]`. Having noothing here, results in the section looping and state being refreshed all the time. Not good, this could be a symptom of badly managed logic idk. Need some thinking brain thoughts to think about and read about useEffect()

Also rendering doesn't seem to wait for `getSubmissions(id)` but it SHOULD, because the call is synchronous, maybe somewhere along the overmind action->grpcManager->method.Call, doesn't actually wait for state at all. Needs more testing.

---

#### Need to add group submissions to state as well!!
Right now submissions doesn't have group submissions in state. 

---

#### Speculative (ignore maybe, doesn't matter too much)
How can we structure calls to state update more healthily. For example going to /home and /course, should both be able to load state. However if you are going from /home to /course. Woiuldn't it be better to have a check and see if there already exists state for this course, making you not need to "refresh" state. However this can create issues such as missing updates that should happen, i.e a push event to github, would require a hard reload to see the changes in state, on the page. It seems like the only option here is to do these updates all the time. But maybe this is how it is meant to function

#### Admin/teacherscopes:
Admin/teacher things haven't been added at all yet.
Haven't really looked at anything like this, but it should be relatively straightforward because of 'has_teacher_scopes' or 'is_admin'.
Might need an additional look at authorization and authentication.

---

##### gRPC server-side streaming:
not implemented at all. on Lab opening issue a stream call. 'maybe a prompt to see if it works'
need to setup repository as well for this to be checked and tested at all.
Conceptually:
```
Callstack concept
ClientSide : [user on lab/id -> grpc stream call to gRPCManager()]

ServerSide : 
web/autograderService.go has a map:
  (userid : <- channel SubmissionStream)
  
on hooks/github.go HandlePush(event,payload, runner)
  Payload.getName()
  Look for stream in map.
  If so, send message on SubmissionStream indicating its handling the push
  run tests.
  Send result (Submission object.) on submissionstream (+end code to indicate to reload the information on the frontend)


```
##### additional Notes gRPC
Only one stream per user, only when they are on a lab page i.e app.com/course/1/assignment2
How to ensure that if user is getting stream from lab1, and the user pushes to lab2, that lab2 results don't gets posted to lab1.


---

Things related to design:

	Overall:
		Links in the navnbar should be more tab-like, can probably be solved with bootstrap for the most part.

	Home page:
		Make it looks cleaner in general, decide on what to have on the home(login landing page).

		LabTable:
			Styling.
			Add css class to TR or TD based on status Not Approved/ Not high enough score etc. and how much time is left before the deadline.

	Course/id page:
			Create a better overview. For assignments and submissions, progressbar etc.


	Course overview:
		create and overviews

	Individual lab Results:
		Styling it better obviously.
		This is the place where we should hook in the gRPC stream request.
		See if we can modify the ouput from the submission/buildinfo to seperate them into their own respective divs


	Info page:
		copy paste and re-do where needed.


