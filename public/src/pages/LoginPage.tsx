import React from "react";

const LoginPage = () => {
    return (
        <div style={{
            display: "flex",
            justifyContent: "center",
            alignItems: "center",
            height: "100vh",
            backgroundColor: "#0d1117",
            color: "#c9d1d9",
            fontFamily: "-apple-system, BlinkMacSystemFont, 'Segoe UI', Helvetica, Arial, sans-serif"
        }}>
            <div style={{
                backgroundColor: "#161b22",
                padding: "40px",
                borderRadius: "10px",
                boxShadow: "0 4px 14px rgba(0, 0, 0, 0.4)",
                textAlign: "center",
                width: "400px"
            }}>
                <img
                    src="https://github.githubassets.com/images/modules/logos_page/GitHub-Mark.png"
                    alt="GitHub Logo"
                    style={{ width: "80px", marginBottom: "20px" }}
                />
                <h2>Sign in to GitHub</h2>
                <p style={{ fontSize: "14px", color: "#8b949e" }}>to continue to QuickFeed</p>

                <a
                    href="/auth/github"
                    className="signIn"
                    style={{
                        display: "block",
                        width: "100%",
                        backgroundColor: "#238636",
                        color: "#fff",
                        textAlign: "center",
                        padding: "10px 15px",
                        borderRadius: "5px",
                        fontSize: "16px",
                        fontWeight: "bold",
                        textDecoration: "none",
                        marginTop: "20px"
                    }}
                >
                    Sign in with GitHub
                    <i className="fa fa-github" style={{ marginLeft: "10px" }}></i>
                </a>
            </div>
        </div>
    );
};

export default LoginPage;
