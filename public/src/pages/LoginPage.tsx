import React, { useEffect } from "react"
import AboutPage from "./AboutPage"
import { nextURL } from "../Helpers"

const LoginPage = () => {
  // Add a class to the body element to style the login page
  useEffect(() => {
    document.body.classList.add("login-page")
    return () => {
      document.body.classList.remove("login-page")
    }
  }, [])
  return (
    <div className="loginContainer">
      <h1 className="loginWelcomeHeader">Welcome to QuickFeed</h1>
      <p className="lead mt-5 mb-5" style={{ textAlign: "center", marginBottom: "50px" }}>
        To get started with QuickFeed, please sign in with your GitHub account.
      </p>
      <section id="loginBox">
        <div className="loginBox">
          <i className="fa fa-5x fa-github align-middle ms-auto" id="github icon" />
          <h4>Sign in with GitHub</h4>
          <p className="text-secondary"> to continue to QuickFeed </p>
          <a href={`/auth/github?next=${nextURL()}`} className="loginButton"> Sign in </a>
        </div>
      </section>
      <AboutPage />
    </div>
  )
}

export default LoginPage
