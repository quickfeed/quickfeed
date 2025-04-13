import React from "react"

type FormProps = {
    prepend: string,
    name: string,
    onChange: (e: React.ChangeEvent<HTMLSelectElement>) => void,
    options: {
        value: string,
        key: string,
        text: string,
    }[],
}

const FormSelect = ({ prepend, name, onChange, options }: FormProps) => {
    return (
        <div className="input-group mb-3">
            <div className="input-group-prepend">
                <div className="input-group-text">{prepend}</div>
            </div>
            <select name={name} onChange={onChange} className="form-control form-select">
                {options.map((option) =>
                    <option value={option.value} key={option.key}>{option.text}</option>
                )}
            </select>
        </div>
    )
}

export default FormSelect
