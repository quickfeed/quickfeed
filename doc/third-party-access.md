# Accessing QuickFeed with third-party applications

## Creating a QuickFeed User

First you will have to create a user on QuickFeed if you do not already have one. You can do this by logging in to QuickFeed with your GitHub account. Your application will be granted the same access permissions as your user.

## Creating a Personal Access Token

To access QuickFeed with your application you need will need to create a GitHub personal access token. You need to use the same GitHub account as the one you used to create your QuickFeed user. You can create a personal access token by following the instructions [here](https://docs.github.com/en/github/authenticating-to-github/creating-a-personal-access-token).
You can use either a [classic personal access token](https://github.com/settings/tokens/new), or a [fine-grained personal access token](https://github.com/settings/personal-access-tokens/new).

The token you create should **not** have any permissions granted. QuickFeed will only use the token to identify you.

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
    "name":"Test User", 
    "studentID":"123456", 
    "email":"test.user@example.com", 
    "avatarURL":"https://example.com/avatar.png", 
    "login":"TestUser", 
}
```
