# AI Agent Instructions for QuickFeed Development

QuickFeed is a Go/TypeScript web application for automated feedback on programming assignments.
It features a Go backend with gRPC/Connect services and a React/TypeScript frontend with Overmind state management.
This file focuses on helping AI agents understand how to develop QuickFeed effectively, with emphasis on code quality, testing patterns, and architectural understanding rather than deployment procedures.

## Code Style Guidelines

### Go Code Style

Ensure that the length of functions does not compromise cyclomatic complexity and readability.
Check using `golangci-lint`.
Aim for a cyclomatic complexity of 15 or less per function.
Keep functions focused and break down complex logic into smaller, well-named helper functions.

Follow the standard Go formatting conventions.
Always run `gofumpt` before committing your changes to ensure consistent formatting.

Follow [Google Go style guidelines](https://google.github.io/styleguide/go/index) for writing clear and maintainable code.
Use idiomatic Go practices and conventions to ensure consistency across the codebase.

Add helper methods to protocol buffer message types in the `qf` and `score` packages to make code more readable and maintainable.

When writing Go tests, use the `testing` package and follow the standard Go testing conventions, including table-driven tests where appropriate.

### Frontend Code Style

When designing frontend features, think critically about the user experience and how to make the interface intuitive and efficient.
Use as few clicks as possible to achieve a task.

Make each page self-explanatory through clear labels, placeholders, and tooltips.
When a page needs more explanation than that, put it in the page, next to the control it explains.

Follow TypeScript best practices and maintain type safety throughout the codebase.
Use proper interfaces and type definitions for all data structures.

### Comments

These rules apply to Go, TypeScript, and `.proto` files alike.

Write doc comments in succinct Go style: a sentence or two, starting with the identifier's name, that state the end result for the caller, such as what a function returns or how the state of the system changes.
Document only what the name and signature leave out, such as a precondition, an error the caller must handle, or a side effect.
Avoid redundant information that is already clear from the code.
Only write comments when the code itself is not clear enough to convey the intent or behavior.

Edit a doc comment only when the behavior it states changes.
A change to the implementation beneath it leaves the comment as it is.

### Documentation Style

When writing documentation in markdown files, ensure proper formatting and structure that follows formatting and style guidelines of the markdown linter.
Follow the one sentence per line rule for better readability and version control diffs.

The markdown files in `doc/` cover what neither the code nor the UI can tell their reader: deploying, configuring, and developing QuickFeed, and what a teacher sets up outside QuickFeed, such as the course organization, its repositories, and test scripts.
Update them only when what their reader must do changes, such as a new setting or a new deployment step.
Keep implementation details in the code: limits, defaults, storage formats, and how a feature works inside.
Explanations of a page's controls belong in the frontend (see Frontend Code Style), not in `doc/teacher.md`.

## Development Workflow

### Commit Strategy

When working on a larger task, make sure to create smaller commits for each logical change.
This makes it easier to review and understand the changes.

Each commit should represent a single logical unit of work that can be easily reviewed and potentially reverted if needed.

Write clear, descriptive commit messages that explain what was changed and why.

### Testing Requirements

Always add tests for RPC service methods using the `web.MockClient()` test helper.
Study the existing tests in the web package to understand how to use this helper effectively.

The MockClient test helper allows you to simulate RPC calls and assert the expected behavior of your service methods without relying on a real backend.

Example MockClient usage pattern:

```go
func TestMyRPCMethod(t *testing.T) {
    db, cleanup := qtest.TestDB(t)
    defer cleanup()

    client := web.MockClient(t, db, scm.WithMockOrgs("admin"), nil)
    // For authenticated requests:
    // client, tm := web.MockClientWithOption(t, db, scm.WithMockOrgs("admin"))

    // Test your RPC method
    response, err := client.MyRPCMethod(context.Background(), &connect.Request[qf.MyRequest]{
        Msg: &qf.MyRequest{
            // Request parameters
        },
    })

    // Assert expected behavior
    if err != nil {
        t.Error(err)
    }
    // Additional assertions...
}
```

Always test both success and error cases for your RPC methods.
Use `qtest.CheckError()` helper for testing expected error responses.

Write comprehensive tests that cover edge cases and boundary conditions.
Aim for high test coverage, especially for business logic and RPC service methods.

### Frontend Testing

Write Jest tests for React components and TypeScript utilities.
Follow the existing patterns in `public/src/__tests__/` directory.

Test user interactions and state management through Overmind actions.
Ensure components render correctly with different props and state configurations.

## Architecture Understanding

### Backend Structure

- `main.go` - Application entry point and server setup
- `qf/` - Protocol buffer definitions and generated Go code for APIs
- `web/` - HTTP handlers, RPC service implementations, and authentication
- `database/` - Database models, queries, and data access layer
- `internal/` - Internal packages for configuration, utilities, and helpers
- `scm/` - Source control management integration (GitHub, GitLab)
- `ci/` - Continuous integration and assignment testing logic

### Frontend Structure

- `public/src/` - TypeScript/React source code
- `public/src/overmind/` - State management with Overmind
- `public/src/components/` - Reusable React components
- `public/src/pages/` - Page-level components and routing
- `public/dist/` - Generated build artifacts (do not edit)

### Key Development Patterns

#### Protocol Buffer Workflow

When editing protocol buffers in `qf/*.proto`:

1. Run `make proto` to regenerate Go and TypeScript code
2. Update affected Go service methods in `web/` package
3. Update frontend TypeScript code to use new types
4. Add comprehensive tests for new RPC methods

#### Request and Response Message Types

QuickFeed does not maintain separate message types for transport and storage.
The types in `qf/types.proto` are both the API messages and the GORM models, and
RPC methods take these types directly:

```proto
rpc CreateGroup(Group) returns (Group) {}
rpc CreateEnrollment(Enrollment) returns (Void) {}
rpc CreateAssignmentFeedback(AssignmentFeedback) returns (Void) {}
```

Define a dedicated `*Request` message in `qf/requests.proto` only when one of the
following applies:

1. The request is a query rather than an entity, such as `SubmissionRequest`,
   `GroupRequest`, `EnrollmentRequest`, or `CourseRequest`.
2. The entity does not carry the `CourseID` that the access control interceptor
   needs. `ReviewRequest` exists for this reason alone: `Review` has no
   `CourseID`, so `checkTeacher` could not determine the course from it.

The interceptor resolves the course through the duck-typed `courseIDProvider`
interface in `web/interceptor/provider_interface.go`, so any type with a
`GetCourseID()` method works as a request without a wrapper. Wrapping an entity
that already has a `CourseID` adds a type and a level of nesting without
enabling anything.

Two consequences follow from types serving both roles:

- Do not use `oneof` in `qf/types.proto`. A `oneof` generates an interface-typed
  field that GORM cannot map to columns. Express exactly-one-of semantics as
  plain fields with a comment, as `Submission` does with `userID` and `groupID`.
  `oneof` is fine in `qf/requests.proto`, where the messages are not persisted.
- Handlers must treat client-supplied values for server-owned fields as
  untrusted. Clear or overwrite IDs, author or owner references, and timestamps
  before persisting; never pass the request message to the database unchanged.

Each request type should implement `IsValid()`, which the validation interceptor
calls before the handler runs. Put invariants that hold for every RPC using the
type in `IsValid()`, and keep per-RPC checks in the handler. A create and an
update share one type but differ on whether the ID must be set, so that check
belongs in the handler.

#### RPC Service Development

1. Define the RPC method in appropriate `.proto` file
2. Implement the method in the corresponding `web/` service file
3. Add comprehensive tests using `MockClient` test helper
4. Update frontend client code to call the new RPC method
5. Add frontend tests for the new functionality

#### Database Changes

When modifying database models or queries:

1. Update the model structs in `database/` package
2. Add database migration if schema changes are needed
3. Update related RPC service methods
4. Add tests that verify database operations work correctly
5. Ensure backwards compatibility where possible

## Code Quality Standards

### Before Finishing a Change

Always run these commands before finishing a change:

1. `go fix ./...` - Apply recommended Go fixes
2. `golangci-lint run ./...` - Run linter to catch potential issues
3. `gofumpt -w .` - Format Go code consistently
4. `cd public && npm run lint` - Check frontend code style
5. `make test` - Run complete test suite to ensure nothing is broken
6. `git diff` - Review your changes carefully before committing

### Error Handling

Always handle errors appropriately in Go code.
Use the `connect` package error types for RPC methods:

```go
if err != nil {
    return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to process request: %w", err))
}
```

Provide meaningful error messages that help users understand what went wrong.

### Performance Considerations

Be mindful of database query performance, especially for operations that may involve large datasets.
Use appropriate database indexes and consider query optimization.

For frontend code, avoid unnecessary re-renders and optimize component performance where needed.
Use React best practices for state management and component lifecycle.

## Quick Reference

### Essential Build Commands

```bash
make download        # Download Go dependencies (~20 seconds)
make install         # Build Go backend (~52 seconds)
make ui             # Build frontend (~4.5 seconds)
make test           # Run all tests (~93 seconds)
```

### Development Server

```bash
# Setup (one-time)
cp .env-template .env
# Edit .env for localhost development

# Start development server
PORT=8080 quickfeed -dev
```

### Testing Specific Areas

```bash
go test ./web/...           # Test web services
go test ./database/...      # Test database layer
cd public && npm run test   # Test frontend
```

### Code Formatting

```bash
gofumpt -w .               # Format Go code
cd public && npm run lint  # Check frontend style
```
