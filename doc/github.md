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
    2. Homepage URL should be your public landing page for QuickFeed, e.g.,

       ```url
       https://uis.itest.run/
       ```

    3. Authorization callback URL must be unique for each instance of QuickFeed, e.g.,

       ```url
       https://uis.itest.run/auth/callback/
       ```

5. Click Register application, and save your Application ID, Client ID, and Client Secret.

6. In the `Private keys` section, click `Generate private key` and save the generated key to the file `$QUICKFEED/internal/config/github/quickfeed.pem`.

7. Save the App ID, Client ID, Client Secret, and path to the saved key in environment variables in your `$QUICKFEED/.env` file, as shown below:

   ```sh
   QUICKFEED_APP_ID="Application ID"
   QUICKFEED_APP_KEY=$QUICKFEED/internal/config/github/quickfeed.pem
   QUICKFEED_CLIENT_ID="Client ID"
   QUICKFEED_CLIENT_SECRET="Client secret"
   ```

8. A public link will be generated on the application's main page.
   This link allows users to install this application on their organizations.
   Installing the application creates an installation that can access the organization according to the application's permissions.

9. Permissions can be set up when the application is created or later, in the `Permissions & events` tab.
   QuickFeed requires the following set of permissions:

   - Repository permissions:
      - Administration: read & write
      - Contents: read & write
      - Issues: read & write
      - Pull requests: read & write
   - Organization permissions:
      - Administration: read & write
      - Members: read & write

   Note that changing the permissions later will not update them automatically for existing installations.
   Organization owners will need to accept changes manually in the organization settings.

10. Start the QuickFeed server:

   ```sh
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
