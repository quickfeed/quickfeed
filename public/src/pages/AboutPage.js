import React from "react";
import Alerts from "../components/alerts/Alerts";
import FeatureBlock, { MiniFeatureBlock } from "../components/FeatureBlock";
import BackToTop from "../components/BackToTop";
const AboutPage = () => {
    return (React.createElement("div", null,
        React.createElement(Alerts, null),
        React.createElement("div", { className: "container" },
            React.createElement("h2", { className: "featurette-heading mt-5" }, "About QuickFeed"),
            React.createElement("p", { className: "lead" }, "QuickFeed is a tool for providing automated feedback to students on their lab assignments. QuickFeed builds upon version control systems and continuous integration. When students upload code to their repositories, QuickFeed automatically builds their code and provides feedback based on tests supplied by the teaching staff. When grading assignments, teaching staff can access the results of test execution and have a valuable tool in the grading process."),
            React.createElement("hr", { className: "loginDivider" }),
            React.createElement("div", { key: "rowheader", className: "row marketing" },
                React.createElement(MiniFeatureBlock, { title: "GitHub Integration", content: "Manage all students and courses on GitHub. Each student gets their own repository. Teachers get separate repositories for publishing assignments and information to students. All taken care of automatically.", media: React.createElement("i", { className: "fa fa-github", style: { fontSize: "140px" } }) }),
                React.createElement(MiniFeatureBlock, { title: "Continuous Integration", content: "Instantaneous feedback to students on how well their code performs. Students can quickly identify what they need to focus on to improve. All customizable for the teaching staff.", media: imageMedia("/assets/img/overlapping-arrows-no-background.webp", "Continuous Integration") }),
                React.createElement(MiniFeatureBlock, { title: "Fair Grading", content: "On due date of an assignment, the most recent version of each student's code is available through GitHub. Easily accessible for the teachers. Together with latest build log, this makes grading easier and more fair.", media: imageMedia("/assets/img/Aplus2-no-background.webp", "Fair Grading") })),
            React.createElement("hr", { className: "loginDivider" }),
            React.createElement(FeatureBlock, { heading: "QuickFeed", subheading: "Automated student feedback", content: "QuickFeed aims to provide students with fast feedback on their lab assignments, and is designed to help students learn about state-of-the-art tools used in the industry. QuickFeed builds upon version control systems and continuous integration. When students upload code to their repositories, QuickFeed automatically builds their code and provides feedback based on tests supplied by the teaching staff. When grading assignments, teaching staff can access the results of test execution and have a valuable tool in the grading process.", imageSrc: "/assets/img/intro1.png" }),
            React.createElement("hr", { className: "loginDivider" }),
            React.createElement(FeatureBlock, { heading: "GitHub Integration", subheading: "Managing courses and students", content: "A course is an organization on GitHub. Students get access to their own private GitHub repository. Uploading their code for review or grading, students can learn to use git for version control.", imageSrc: "/assets/img/intro3.png", reverse: true }),
            React.createElement("hr", { className: "loginDivider" }),
            React.createElement(FeatureBlock, { heading: "Continuous Integration", subheading: "Builds and tests student code", content: "As code gets pushed up to GitHub, an automatic build process defined by the teacher, generates feedback to students. When the build process is completed, student gets immediate access to this feedback on their personal course page. Tests defined by either teachers or students will be processed and tell students about their progress on the assignments.", imageSrc: "/assets/img/intro2.png" }),
            React.createElement("hr", { className: "loginDivider" }),
            React.createElement(FeatureBlock, { heading: "Grading", subheading: "Easy and Fair", content: "On the due date, teachers can access the test results and use this as a tool in the grading process. The teaching staff will immediately know which of their tests passed, and how much of the code is covered by the tests.", imageSrc: "/assets/img/intro4.png", reverse: true }),
            React.createElement(BackToTop, null))));
};
function imageMedia(src, alt) {
    return (React.createElement("img", { src: src, alt: alt, style: { width: "100%", height: "100%", objectFit: "contain" } }));
}
export default AboutPage;
