import React from "react"


/** SignupText is used to display an error message if the user information is incomplete. */
const SignupText = () => {
    return (
        <div className="alert alert-error mb-6 shadow-lg">
            <div className="flex flex-col gap-2">
                <div className="flex items-center gap-2">
                    <i className="fa fa-exclamation-circle text-xl"></i>
                    <span className="font-bold">Complete Your Profile</span>
                </div>
                <div className="text-sm space-y-1">
                    <p>Fill in the form below to complete signup.</p>
                    <p>Use your <strong>real name</strong> as it appears on Canvas.</p>
                    <p>If your name does not match any names on Canvas, you will not be granted access.</p>
                </div>
            </div>
        </div>
    )
}

export default SignupText
