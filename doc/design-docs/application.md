# Allowing third party applications to communicate with QuickFeed

We should implement a system that will streamline the process of integrating QuickFeed into separate services, such as the [QuickFeed HelpBot](https://github.com/quickfeed/helpbot), and command-line tools such as [approvelist](https://github.com/quickfeed/quickfeed/blob/master/cmd/approvelist/main.go).

Currently we set an environment variable, `QUICKFEED_AUTH_TOKEN`, and load this into QuickFeed. This token is then associated with a hard-coded user ID. Any incoming requests that includes a token matching `QUICKFEED_AUTH_TOKEN` will be assumed to be the hard-coded user.

## Proposed Solution

### Registering an application

We should allow a given user to register their application with QuickFeed. This process should include:

- Naming their application
- Providing a short description for their intended use
- Provide required scopes for their application

Upon successfully registering, the user should be granted:

- Client ID
- Client Secret

This information should be stored in the database:

```go
message Application {
    uint64 ID           = 1; // Not necessarily needed
    uint64 userID       = 2; // Owner of application
    string clientID     = 3; // Used to identify existing secret 
    string name         = 4;
    string description  = 5;
    string secret       = 6; // Securely stored secret
    // Scopes, which will be encoded into JWT upon authentication
}
```

### Authenticating an application

---

An application should be able to authenticate with QuickFeed with a generated client ID and secret pair.

The application must send a request with gRPC to a specific endpoint, with the aforementioned pair attached in its metadata.

QuickFeed will then do a database lookup using the received client ID. The matching row will contain a salted and hashed version of the client secret. If the provided secret matches the stored one, we will consider the authentication as successful.

### Authorizing an application

---

Following successful authentication, an application must be granted authorization to perform a set of actions. These actions must be defined in the scope when registering an application.

The scope cannot extend beyond the scopes of the owner of the registered application.

### Finalizing the process

---

After authenticating and retrieving application scopes, QuickFeed will respond to the application requesting access. The response will include a signed JWT. Encoded in this JWT is all the information required to process any requests the application is scoped for.

The application **must include this token in all subsequent requests** made to the QuickFeed server.

### Managing registered applications

---

The frontend should provide all users with the means to view any existing applications belonging to themselves. We should provide the user with the means to create, view, and delete applications.

## Further details

### Required RPCs

```go
// Alternatively use Application message in both request and response
rpc CreateApplication(ApplicationRequest) returns (ApplicationResponse) {}
```

The `CreateApplication` RPC will return the generated client ID and client secret. The client secret will only ever be returned once, and subsequent visits to the frontend to manage the application will **only** list the client ID. The user **must save** the client secret when received.

```go
rpc DeleteApplication(Application) returns (Void) {}
```

The `DeleteApplication` RPC will delete an application if the `Application` message contains an ID (or client ID) that exists and belongs to the authenticated user.

```go
rpc GetApplications(Void) returns (Applications) {}
```

The `GetApplications` RPC will return all applications registered to the authenticated user. The returned applications will include:

- Name
- Description
- Client ID

```go
rpc AuthenticateApplication(Void) returns (Void) {}
```

The `AuthenticateApplication` RPC will be used to authenticate an application. The request must include both a client ID and a client secret in its metadata. If the request is successful, the response must include a signed JWT. If the response does not include a signed JWT, the request was not successful.

Alternatively, we could pass this information back and forth using defined messages.

### Client ID

The client ID is used to identify the application owner and to retrieve any existing client secrets. Only one database row should ever match a given client ID.

The client ID should be sufficiently unguessable, but does not necessarily have to be hashed.

### Client Secret

> Note: I'm not entirely sure which method is best to use to store the secret safely.

The generated client secret should be unique and sufficiently unguessable. It must be stored securely in the database.

We could generate a UUID, salt and hash this, and store in the database.

When the server receives an `AuthenticateApplication` request, we can compare the received plaintext secret with the hashed secret in our database.

### Scopes

> TODO
