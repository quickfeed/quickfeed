import React from 'react'

export enum KnownMessage {
    NoSubmission = "Select a submission from the results table",
    NoAssignment = "Assignment does not have a submission",
}

// CenteredMessage is a component that displays a message in the center of the screen
// Can be used as a placeholder and to inform the user.
export const CenteredMessage = ({ message }: { message: string | KnownMessage }) => {
    return <div className="text-center mt-5"><h3>{message}</h3></div>
}
