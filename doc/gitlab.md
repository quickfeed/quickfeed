# GitLab Setup

TODO(meling) These instructions are outdated, and would require updating the `scm` package.

1. The first step is to create a new API key for your Autograde application.
    1. Log into your gitlab account.
    2. Click your profile picture and select Settings.
    3. In the settings, click on Applications.
        1. Name your application I.e Autograde-development
    5. Redirect URL is very important, and should be individual for each unique instance of Autograde.

        <https://{base-url}/auth/gitlab/callback>
    6. Once you click create, you will be able to retrieve your
        1. Application-ID (called client-ID in Autograde)
        2. Secret (called client-secret in Autograde)

        These has to be set as environment variables.
        see. <a href="Installation.md"> Installation manual </a>

2. Second step is to create a Group.
    1. Log into www.gitlab.com
    2. Top right press the + button and click on "New group"

    3. Name your group

    4. Go into Organization settings.
    5. You should checkout the billing plans to see which one fits your requirements.

        Note: <br/>
        Group-Webhooks are not part of the free plan! Only Per-Project (not currently supported) <br/>
        <a href="https://about.gitlab.com/pricing/gitlab-com/feature-comparison/"> See here </a>

    6. Select the Webhook menu inside setting page.

        1. Add Autograde hook URL.

                https://{baseurl}/hook/gitlab/events

        2. Create a secret.
        3. Make sure The push event trigger is selected.
        4. If you want SSL, make sure Enable SSL validation is selected.
        5. Create the webhook.
