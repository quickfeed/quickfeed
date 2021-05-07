import React from "react";


export const DynamicForm = () => {
    const print = (text: string) => {
        console.log(text)
    }
    let elements: {[key: string]: {placeholder: string, func: Function, type?: string}} = {
        "username": {placeholder: "Your Username", func: print}, 
        "email": {placeholder:"Your Email", func: print}, 
        "password": {placeholder:"Your Password", func: console.log}
    }

    let InputElements: JSX.Element[] = []
    for (const key in elements) {
        InputElements.push(
            <React.Fragment>
            <label htmlFor={key}>{key}</label>
            <input className="form-control" name={key} type="text" placeholder={elements[key].placeholder} onClick={e => elements[key].func(e.currentTarget.value)}></input>
            </React.Fragment>
        )
    }

    return (
        <div className="box">
            <div className="jumbotron"><div className="centerblock container"><h1></h1>You can edit your user information here.</div></div>
            <form className="form-group well" style={{width: "400px"}} onSubmit={e => {e.preventDefault(); submitHandler()}}>
                        {InputElements}
            </form>
        </div>
    )
}

export default DynamicForm

function submitHandler() {
    throw new Error("Function not implemented.");
}
