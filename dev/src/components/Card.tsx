import React from "react"
import { useHistory } from "react-router"


export interface Notification {
    color: string,
    text: string,
}

const Card = (props: {title: string, notification?: Notification, text: string, buttonText: string, to: string}): JSX.Element => {
    const history = useHistory()

    const notification = props.notification ? <i className={`badge badge-${props.notification.color} float-right`}>{props.notification.text}</i> : null

    return (
        <div className="col-sm-6" style={{marginBottom: "10px"}}>
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
                    <div className="btn btn-primary" onClick={() => history.push(props.to)}>
                        {props.buttonText}
                    </div>
                </div>
            </div>
        </div>
    )
}

export default Card