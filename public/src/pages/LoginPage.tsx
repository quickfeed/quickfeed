import React from "react"
import AboutPage from "./AboutPage"

const LoginPage = () => {
    return (
        <div className="min-h-screen flex flex-col items-center bg-base-100 text-base-content">
            <h1 className="text-5xl font-bold mt-16 mb-12 text-base-content">
                Welcome to QuickFeed
            </h1>
            <p className="text-lg text-center mb-12 max-w-2xl px-4 text-base-content/80">
                To get started with QuickFeed, please sign in with your GitHub account.
            </p>
            <section className="mb-12">
                <div className="card bg-base-200 shadow-xl p-8 text-center min-w-[300px]">
                    <i className="fa fa-5x fa-github mb-4 text-base-content" />
                    <h4 className="text-xl font-semibold mb-2">Sign in with GitHub</h4>
                    <p className="text-base-content/60 mb-6">to continue to QuickFeed</p>
                    <a
                        href="/auth/github"
                        className="btn btn-success text-white hover:btn-success/90 transition-colors"
                    >
                        Sign in
                    </a>
                </div>
            </section>
            <AboutPage />
        </div>
    )
}

export default LoginPage
