# GitHub Setup

## Generating API Keys for Connecting to GitHub

The first step is to create a new API key for your QuickFeed application.

1. Decide which GitHub account to use for connecting your QuickFeed server to GitHub.
   - For development, using a private GitHub account is fine.
   - For deployment, we recommend creating a separate GitHub account.
2. Log in to your chosen GitHub account.
3. Click your profile picture and select Settings.
4. In Settings, click Developer Settings and create a [New GitHub App](https://docs.github.com/en/enterprise-cloud@latest/developers/apps/building-github-apps/creating-a-github-app).
    1. Name your application, e.g., `QuickFeed Development`.
    2. Homepage URL should be your public landing page for QuickFeed, e.g., `https://uis.itest.run/`.
    3. Authorization callback URL must be unique for each instance of QuickFeed, e.g., `https://uis.itest.run/auth/callback/`.
5. When you click Register application, you will be able to retrieve your Application ID, Client ID and Client Secret. 
   These are necessary for QuickFeed to access GitHub, and for GitHub to communicate back to the QuickFeed server.
   In `Private keys` section click the `Generate private key` button and save the generated key to a file. The default path for the key file is `quickfeed/internal/config/github/quickfeed.pem`.
   Create a shell script `quickfeed-env.sh` and copy the App ID, Client ID and Client Secret values, and path to the generated key (if you wish to store it outside of the default path) as shown below:

   ```sh
   export QUICKFEED_APP_ID="Application ID"
   export QUICKFEED_APP_KEY="Path to the application's private key"
   export QUICKFEED_CLIENT_ID="Client ID"
   export QUICKFEED_CLIENT_SECRET="Client secret"
   ```

6. A public link will be generated on the application's main page. This link allows users to install this application on their organizations. Installing the application creates an installation that can access the organization according to the application's permissions.
7. Permissions can be set up when the application is created or later, in the `Permissions & events` tab. QuickFeed requires the following set of permissions:


8. Then, to start the QuickFeed server:

   ```sh
   % source quickfeed-env.sh
   % quickfeed -service.url uis.itest.run &> quickfeed.log &
   ```

   The first GitHub user to login becomes admin for the server.

   Note that the `-service.url` should not be specified with the `https://` prefix.

## Creating a Course

To create a course on QuickFeed, you must first create a GitHub organization for your course.

At the University of Stavanger, you can create new organizations under our [enterprise account](https://github.com/enterprises/university-of-stavanger).
If you are employee or teaching assistant at the University of Stavanger, contact Hein Meling for more information.

1. Log into a GitHub account with access to the enterprise account.
2. Navigate to the [enterprise account](https://github.com/enterprises/university-of-stavanger).
3. Click New organization.
4. Name your organization, e.g., `qf101-2020`.
5. Skip Invite members; QuickFeed handles this.

Others may follow the instructions below.

1. Log into the teacher's GitHub account.
2. In the top right corner, press the + menu and click on New organization.
3. Select billing plan.
   Note that QuickFeed requires a billing plan to allows private repositories.
   Academic institutions can get free access through GitHub's [Academic Campus Program](https://education.github.com/schools).
4. Name your organization, e.g., `qf101-2020`.
5. Skip Invite members; QuickFeed handles this.
6. Skip Organization details.
