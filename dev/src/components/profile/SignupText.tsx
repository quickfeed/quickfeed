import React from "react"

/** SignupText is used to display an error message if the user information is incomplete. */
const SignupText = (): JSX.Element => {
    return (
        <blockquote className="blockquote card-body" style={{ color: "red" }}>
            <p>
                Fill in the form below to complete signup.
            </p>
            <p>
                Use your <i>real name</i> as it appears on Canvas.
            </p>
            <p>
                If your name does not match any names on Canvas, you will not be granted access.
            </p>
        </blockquote>
    )
}

export default SignupText
