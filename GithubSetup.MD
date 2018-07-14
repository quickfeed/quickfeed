# Github Setup
1. The first step is to create a new API key for your Autograde application.
    1. Log into your github account.
    2. Click your profile picture and select Settings.
    3. In the settings, click on Developer Settings.
        1. Name your application I.e Autograde-development
    4. Home page url should be your base-url.
    5. Authorization callback URL is very important, and should be individual for each unique instance of Autograde.

        https://{base-url}/auth/github/callback
    6. Once you click create, you will be able to retrieve your
    client-ID and client-secret, which has to be set in your environment
    see. <a href="Installation.md"> Installation manual </a>


2. Second step is to create a organization.
    1. Log into www.github.com
    2. Top right press the + button and click on "New organization"
    
    3. Name your organization, select billing plan

    Note! 
    If selecting free plan, only teams can be secret, thus test repository will be visible to students. </br>For testing/development the free plan can be used, but if you are planning to use this for your course, you should be checkout their billing options and select the one that fits your criteria.
    </br>Also read through GitHub Corporate Terms of Service.

    4. You can skip the invite member part, since Autograde handles this.
    5. You can skip Organization details.
    6. Go into Organization settings.
        1. Select the Webhook menu.
        2. Add webhook url to providers organization found in the organization setting page.

            https://{baseurl}/hook/github/events