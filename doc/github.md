# GitHub Setup

## Generating API Keys for Connecting to GitHub

The first step is to create a new API key for your QuickFeed application.

1. Decide which GitHub account to use for connecting your QuickFeed server to GitHub.
   - For development, using a private GitHub account is fine.
   - For deployment, we recommend creating a separate GitHub account.
2. Log in to your chosen GitHub account, click on your profile picture and select Settings.
3. In Settings, click Developer Settings and create a New OAuth App.
    1. Name your application, e.g. `QuickFeed Development`.
    2. Homepage URL should be your public landing page for QuickFeed, e.g. `https://uis.itest.run/`.
    3. Authorization callback URL must be unique for each instance of QuickFeed, e.g. `https://uis.itest.run/auth/github/callback`.
4. When you click Register application, you will be able to retrieve your Client ID and Client Secret.
   These are necessary for QuickFeed to access GitHub, and for GitHub to communicate back to the QuickFeed server.
   Create a shell script `quickfeed-env.sh` and copy the Client ID and Client Secret values as shown below:

   ```sh
   export GITHUB_KEY="Client ID"
   export GITHUB_SECRET="Client Secret"
   ```

   Then, to start the QuickFeed server:

   ```sh
   source quickfeed-env.sh
   quickfeed -service.url uis.itest.run -database.file ./qf.db -http.addr :3005 &> qf.log &
   ```

## Creating a Course

To create a course on QuickFeed, you must first create a GitHub organization for your course.

At the University of Stavanger, you can create new organizations under our [enterprise account](https://github.com/enterprises/university-of-stavanger).
If you are employee or teaching assistant at the University of Stavanger, contact Hein Meling for more information.

1. Log into a GitHub account with access to the enterprise account.
2. Navigate to the [enterprise account](https://github.com/enterprises/university-of-stavanger).
3. Click New organization.
4. Name your organization, e.g. `qf101-2020`.
5. Skip Invite members; QuickFeed handles this.

Others may follow the instructions below.

1. Log into the teacher's GitHub account.
2. In the top right corner, press the + menu and click on New organization.
3. Select billing plan.

   Note that QuickFeed requires a billing plan to allows private repositories.
   Academic institutions can get free access through GitHub's [Academic Campus Program](https://education.github.com/schools).

4. Name your organization, e.g. `qf101-2020`.
5. Skip Invite members; QuickFeed handles this.
6. Skip Organization details.
