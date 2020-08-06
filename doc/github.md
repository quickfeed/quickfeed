# Github Setup
1. The first step is to create a new API key for your Autograde application.
    1. Log in to your GitHub account, click on your profile picture and select Settings.
    3. In Settings, click Developer Settings and create a New OAuth App.
        1. Name your application, i.e. Autograder-development.
        2. Homepage URL should be your base-url, e.g. `http://ag2.ux.uis.no/`.
        3. Authorization callback URL must be unique for each instance of Autograder, e.g. `https://ag2.ux.uis.no/auth/github/callback`.
    4. When you click Register application, you will be able to retrieve your Client ID and Client Secret, which are necessary for Autograder to access GitHub, and for GitHub to communicate back to the Autograder server. See the <a href="Installation.md"> Installation manual </a> for details on how to configure the Client ID and Client Secret environment variables.

2. Second step is to create an organization.
    1. Log in to your GitHub account.
    2. In the top right corner, press the + menu and click on New organization.
    3. Name your organization, select billing plan.

       **Note!** If you select the free plan, only teams can be secret, and thus the `tests` repository will be visible to everyone. This should be fine for development and testing, but if you are planning to use Autograder in your course, you will probably want to upgrade your organization to a plan with private repositories. Academic institutions will most probably get free access through GitHub's academic program.

    4. Skip Invite members; Autograder handles this.
    5. Skip Organization details.
    6. There is no need to set up a webhook for organizations anymore, it will be added automatically on course creation.
