import React from "react"
import { useHistory } from "react-router"
import { Color } from "../Helpers"


export interface Notification {
    color: Color,
    text: string,
}

/**  This component displays a card with a header, a body and a button in the footer
 * @param title: The title of the card
 * @param text: The text in body of the card 
 * @param buttonText: The text of the button in the footer
 * @param notification: (optional) Notification to display. Floats right of title. 
 * @param to: (Optional) The path to navigate to when the button is clicked 
 * @param onclick: (Optional) The function to call when the button is clicked 
 */
const Card = (props: { title: string, text: string, buttonText: string, notification?: Notification,  to?: string, onclick?: () => void }): JSX.Element => {
    const history = useHistory()

    const notification = props.notification ? <i className={`badge badge-${props.notification.color} float-right`}>{props.notification.text}</i> : null

    // TODO: Maybe support both onclick and to, rather than having to choose one. Not sure where it would be used though.
    /* Runs the provided function, or redirect, on button click. If both 'to' and 'onclick' are provided, 'onclick' takes precedence */
    const onClick = () => {
        if (props.onclick) {
            props.onclick()
        } else if (props.to) {
            history.push(props.to) // Redirect to the given URL
        }
    }
    return (
        <div className="col-sm-6" style={{ marginBottom: "10px" }}>
            <div className="card">
                <div className="card-body">
                    <h5 className="card-title">
                        {props.title}
                        {" "}
                        {notification}
                    </h5>
                    <p className="card-text">
                        {props.text}
                    </p>
                    <div className="btn btn-primary" onClick={onClick}>
                        {props.buttonText}
                    </div>
                </div>
            </div>
        </div>
    )
}

export default Card
