# Accessing QuickFeed with Third-party Applications

## Creating a QuickFeed User

First you will have to create a user on QuickFeed if you do not already have one.
You can do this by logging in to QuickFeed with your GitHub account.
Your application will be granted the same access permissions as your user.

## Creating a Personal Access Token

To access QuickFeed with your application you need will need to create a GitHub personal access token.
You need to use the same GitHub account as the one you used to create your QuickFeed user.
You can create a personal access token by following the instructions [here](https://docs.github.com/en/github/authenticating-to-github/creating-a-personal-access-token).
You can use either a [classic personal access token](https://github.com/settings/tokens/new), or a [fine-grained personal access token](https://github.com/settings/personal-access-tokens/new).

The token you create should **not** have any permissions granted.
QuickFeed will only use the token to identify you.

The personal access token serves as your credentials for accessing QuickFeed.

## Usage

For your application to be granted access to QuickFeed you need to provide the personal access token in the `Authorization` header of your requests.
The token needs to be sent with every request you send to QuickFeed.

Example curl request:

```bash
curl \
    --header 'Content-Type: application/json' \
    --header 'Authorization: <TOKEN>' \
    --data '{}' \
    <https://<YOUR QUICKFEED DOMAIN>/qf.QuickFeedService/GetUser>
```

This should return a response similar to this:

```bash
{
    "ID":"1",
    "Name":"Test User",
    "StudentID":"123456",
    "Email":"test.user@example.com",
    "AvatarURL":"https://example.com/avatar.png",
    "Login":"TestUser",
}
```

## Using the QuickFeedService API

The `QuickFeedService` API is defined in [qf.proto](../qf/quickfeed.proto).
To use the API you need to write a client for your language of choice.
To invoke one of the API methods you need to add the `Authorization` header to your request.
This can be done using a token auth client interceptor, as shown in the following Go example:

```go
func NewQuickFeed(serverURL, token string) qfconnect.QuickFeedServiceClient {
	return qfconnect.NewQuickFeedServiceClient(
		http.DefaultClient,
		serverURL,
		connect.WithInterceptors(
			interceptor.NewTokenAuthClientInterceptor(token),
		),
	)
}
```

See also the [approvelist](../cmd/approvelist/main.go) command for an example of how to use the API.
