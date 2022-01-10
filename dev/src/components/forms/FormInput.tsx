import React from "react"

type FormProps = {
    prepend: string, 
    name: string, 
    placeholder?: string, 
    defaultValue: string | undefined, 
    onChange?: (e: React.FormEvent<HTMLInputElement>) => void, 
    type?: string
}

const FormInput = ({ prepend, name, placeholder, defaultValue, onChange, type }: FormProps): JSX.Element => {
    return (
        <div className="col input-group mb-3">
            <div className="input-group-prepend">
                <div className="input-group-text">{prepend}</div>
            </div>
            <input className="form-control"
                name={name}
                type={type ? type : "text"}
                placeholder={placeholder}
                defaultValue={defaultValue}
                onChange={onChange}
            />
        </div>
    )
}

export default FormInput
