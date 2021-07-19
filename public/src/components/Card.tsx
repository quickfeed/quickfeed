import React from "react"
import { useHistory } from "react-router"


const Card = (props: {title: string, text: string, buttonText: string, to: string}) => {
    const history = useHistory()

    return (
        <div className="col-sm-6">
            <div className="card">
                <div className="card-body">
                    <h5 className="card-title">
                        {props.title}
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